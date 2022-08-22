// Promise functionality for go. There are a couple of Go promise libraries around, this one attempts to balance
// matching how promises typically work, with still being in a "go way".
package promise

import (
	"os"
	"sync"
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
	wg       sync.WaitGroup
	mu       sync.Mutex
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
	p.wg.Add(1)
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

// Returns a promise that resolves with the provided value
func Resolve[T any](val T) *Promise[T] {
	return &Promise[T]{
		value:    val,
		finished: true,
	}
}

// Returns a promise that rejects with the provided error
func Reject[T any](err error) *Promise[T] {
	return &Promise[T]{
		err:      err,
		finished: true,
	}
}

// Waits for the promise to finish, and returns the value and error from the promise.
func (p *Promise[T]) Await() (T, error) {
	p.wg.Wait()
	return p.value, p.err
}

// Calls the supplied function when the promise has resolved, and returns a promise that will resolve when
// the supplied function finishes (allowing for chaining). Note: There is no Catch(), it doesn't really align
// with how go works.
func (p *Promise[T]) Then(fn func(T, error) (T, error)) *Promise[T] {
	next := New(func() (T, error) {
		p.wg.Wait()
		return fn(p.value, p.err)
	})
	return next
}

type result[T any] struct {
	val T
	err error
	idx int
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
		ch := make(chan result[T])

		for idx, p := range promises {
			go func(idx int, p *Promise[T]) {
				v, err := p.Await()
				ch <- result[T]{val: v, err: err, idx: idx}
			}(idx, p)
		}
		res := <-ch
		return res.val, res.err
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
		chResults := make(chan result[T])
		for idx, p := range promises {
			go func(idx int, p *Promise[T]) {
				v, err := p.Await()
				chResults <- result[T]{
					val: v,
					err: err,
					idx: idx,
				}
			}(idx, p)
		}
		results := make([]T, len(promises))
		for i := 0; i < len(promises); i++ {
			res := <-chResults
			if res.err != nil {
				return nil, res.err
			}
			results[res.idx] = res.val
		}
		return results, nil
	})
}

// Returns a promise that will error with os.ErrDeadlineExceeded when the supplied duration elapses.
// This can be combined with promise.Race to run a function that times out. However be cautious as the
// other function will still keep running even after the Race has returned the timeout.
func Timeout[T any](duration time.Duration) *Promise[T] {
	return New(func() (T, error) {
		var t T
		time.Sleep(duration)
		return t, os.ErrDeadlineExceeded
	})
}

func (p *Promise[T]) resolve(v T) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.finished {
		return
	}
	p.value = v
	p.finished = true
	p.wg.Done()
}
func (p *Promise[T]) reject(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.finished {
		return
	}
	p.err = err
	p.finished = true
	p.wg.Done()
}
