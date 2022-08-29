// Promise functionality for go. There are a couple of Go promise libraries around, this one attempts to balance
// matching how promises typically work, with still being in a "go way".
package promise

import (
	"fmt"
	"os"
	"time"
)

// A promise will execute immediately, and the result of the promise (the returned value or error) can be
// retrieved from the promise with Await() or Then() functions. Promises can take some of the leg work out
// of creating a go routine and getting the value back later on. You can pass a promise around outside the
// scope of the original function and later get the value with Await(). Promises are thread safe.
type Promise[T any] struct {
	value    T
	err      error
	finished bool
	done     chan struct{}
}

var closedChan = make(chan struct{})

func init() {
	close(closedChan)
}

// Returns a new promise, that will resolve when the supplied function returns, either the value or an error.
// While this is a generic function, taking a type, go will automatically work out the type based on the return
// type of your supplied function (see example).
//
// Note: Unlike javascript promises, the callback function is not supplied a resolve() and reject() function
// to call. This is because those don't make sense with the way golang works. When it is time to resolve the
// promise, simply return.
//
// Example:
//    p := promise.New(func() (string, error) {
//	      return "Hello world", nil
//    })
//    greeting, _ := p.Await()
//    fmt.Println(greeting) // Outputs: Hello world
func New[T any](fn func() (T, error)) *Promise[T] {
	p := Promise[T]{}
	p.done = make(chan struct{})
	go func() {
		v, err := fn()
		if err != nil {
			p.reject(err)
		} else {
			p.resolve(v)
		}
	}()
	return &p
}

// Returns a promise that resolves with the provided value.
func Resolve[T any](val T) *Promise[T] {
	return &Promise[T]{
		value:    val,
		finished: true,
		done:     closedChan,
	}
}

// Returns a promise that rejects with the provided error.
// When creating a Reject promise, you will need to provide the promise type, as it cannot be
// infered as it is everywhere else.
//
//  // Returns a string promise that errors immediately
//  p := promise.Reject[string](errors.New("Something went wrong"))
func Reject[T any](err error) *Promise[T] {
	return &Promise[T]{
		err:      err,
		finished: true,
		done:     closedChan,
	}
}

// Waits for the promise to finish, and returns the value and error from the promise.
func (p *Promise[T]) Await() (T, error) {
	<-p.Done()
	return p.value, p.err
}

// Calls the supplied function when the promise has resolved, and returns a promise that will resolve when
// the supplied function finishes (allowing for chaining). Note: There is no Catch(), it doesn't really align
// with how go works.
func (p *Promise[T]) Then(fn func(T, error) (T, error)) *Promise[T] {
	next := New(func() (T, error) {
		<-p.Done()
		return fn(p.value, p.err)
	})
	return next
}

// Done returns a channel that's closed when the promise has resolved or rejected. Successive calls to Done return the
// same value, and calling Done on a returned promise will return an immediately closed channel. This is the best
// wait to wait for a promise to resolve without calling Await.
//
//   func myFunc(ctx context.Context) int {
//   	p.New(func() (int, error) {
//   		// Do work...
//   		return 42, nil
//   	})
//   	select {
//   	case <- p.Done():
//   		return p.Await() // Will return immediately
//   	case <- ctx.Done():
//   		// If the ctx is cancelled/timesout, then we return immediately and forget about the promise.
//   		return 0, ctx.Err()
//   	}
//   }
func (p *Promise[T]) Done() chan struct{} {
	return p.done
}

func (p *Promise[T]) String() string {
	return fmt.Sprintf("Promise.%T", p.value)
}

// Returns a promise that resolves to the first value from the supplied promises.
//
// Example:
//    a := promise.New(func() (int, error) { return 10, nil })
//    b := promise.New(func() (int, error) {
//        time.Sleep(time.Second); return 20, nil
//    })
//    val, _ := promise.Race(a, b).Await()
func Race[T any](promises ...*Promise[T]) *Promise[T] {
	return New(func() (T, error) {
		ch := make(chan int, len(promises))

		for idx, p := range promises {
			go func(idx int, p *Promise[T]) {
				<-p.Done()
				ch <- idx
			}(idx, p)
		}
		idx := <-ch
		return promises[idx].value, promises[idx].err
	})
}

// Returns a promise that resolves when all the supplied promises have resolved. The returned promise's value type
// is an array of the promise value types. If any promise returns an error then the returned promise rejects
// immediately. The return order in the array matches the promise order supplied.
//
// Example:
//    all := promise.All(promise.Resolve(10), promise.Resolve(20), promise.Resolve(30))
//    vals, _ := all.Await()
//    fmt.Println(vals) // Outputs: [10 20 30]
func All[T any](promises ...*Promise[T]) *Promise[[]T] {
	return New(func() ([]T, error) {
		results := make([]T, len(promises))
		promiseIdx := make(chan int, len(promises))
		// Spawn multiple go routines to wait on each promise, and then pass the idx through the return channel
		for idx, p := range promises {
			go func(idx int, p *Promise[T]) {
				<-p.Done()
				promiseIdx <- idx
			}(idx, p)
		}
		// Wait for all the promises to resolve.
		for i := 0; i < len(promises); i++ {
			idx := <-promiseIdx // promiseIdx returns the index of the promise that just resolved
			v, err := promises[idx].Await()
			if err != nil {
				return nil, err
			}
			results[idx] = v
		}
		close(promiseIdx)
		return results, nil
	})
}

// Returns a promise that will error with os.ErrDeadlineExceeded when the supplied duration elapses.
// This can be combined with promise.Race to run a function that times out. However be cautious as the
// other function will still keep running even after the Race has returned the timeout.
//
// This function, like promise.Reject, will need to specify the promise type:
//   promise.Timeout[float64](time.Second * 5)
//
// See Dome() for a channel that is a better way to do this, especially with contexts.
func Timeout[T any](duration time.Duration) *Promise[T] {
	return New(func() (T, error) {
		var t T
		time.Sleep(duration)
		return t, os.ErrDeadlineExceeded
	})
}

// Internal functions that resolve/reject the promises

func (p *Promise[T]) resolve(v T) {
	if p.finished {
		return
	}
	p.value = v
	p.finished = true
	close(p.done)
}
func (p *Promise[T]) reject(err error) {
	if p.finished {
		return
	}
	p.err = err
	p.finished = true
	close(p.done)
}
