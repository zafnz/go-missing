package promise_test

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zafnz/go-missing/promise"
)

func ExamplePromise_Then() {
	fortyTwo := promise.New(func() (int, error) {
		time.Sleep(10 * time.Millisecond)
		return 42, nil
	})
	bigger := fortyTwo.Then(func(num int, err error) (int, error) {
		time.Sleep(20 * time.Millisecond)
		return num * 10, nil
	})
	// We now have bigger as a promise, and when we Await on bigger, it will finally
	// return when both it and the original promise have returned (about 30ms later)
	num, _ := bigger.Await()
	fmt.Println(num) // Output: 420
}
func ExamplePromise_Then_chaining() {
	// We can also chain:
	finalPromise := promise.New(func() (string, error) {
		return "Hello", nil
	}).Then(func(str string, err error) (string, error) {
		return str + " World", nil
	}).Then(func(str string, err error) (string, error) {
		return str + "!", nil
	})
	// Note: The way golang works, there's not too much point to chaining like this
	// but this functionality is here, just in case.

	result, _ := finalPromise.Await()
	fmt.Println(result) // Output: Hello World!
}

func ExamplePromise() {
	x := promise.New(func() (int, error) {
		return 42, nil
	})

	// Await pauses execution
	v, _ := x.Await()
	fmt.Println(v)

	// A Then() can be attached at any point, and will call the supplied function with the value (or error)
	// when it becomes available
	x.Then(func(val int, err error) (int, error) {
		fmt.Println(val)
		return 0, nil
	})
	time.Sleep(100 * time.Millisecond) // Give us time for the Then to execute

	// Await will return immediately the second time, as the promise has been resolved.
	v, _ = x.Await()
	fmt.Println(v)
	// Output:
	// 42
	// 42
	// 42
}

func ExamplePromise_errors() {
	// Errors in promises are passed to any Then() functions and as a return in Await()
	p := promise.New(func() (int, error) {
		return 0, errors.New("something went wrong")
	})
	_, err := p.Await()
	if err != nil {
		fmt.Println(err.Error())
	}
	p.Then(func(i int, err error) (int, error) {
		if err != nil {
			fmt.Println("Inside Then(), we see the promise got an error!")
		}
		return 0, err // We can propegate the error if we want, or return another value instead
	})
	// For the Example purpose we need to wait a few milliseconds for that .Then() output
	// to occur.
	time.Sleep(100 * time.Millisecond)
	// Output:
	// something went wrong
	// Inside Then(), we see the promise got an error!
}

func TestPromiseChain(t *testing.T) {
	var stage int
	x := promise.New(func() (int, error) {
		if stage != 0 {
			t.Fatal("First promise not at stage 0")
		}
		stage++
		return 42, nil
	})
	y := x.Then(func(x int, err error) (int, error) {
		if stage != 1 {
			t.Fatal("First Then() in chain not at stage 1")
		}
		stage++
		if x != 42 {
			t.Fatal("First Then() val not 42")
		}
		return x + 10, nil
	}).Then(func(x int, err error) (int, error) {
		if stage != 2 {
			t.Fatal("Second Then() in chain not at stage 2")
		}
		stage++
		if x != 52 {
			t.Fatal("First Then() val not 52")
		}
		return x + 100, nil
	})
	v, _ := y.Await()
	if v != 152 {
		t.Errorf("Final value not 152: %d", v)
	}
}

func TestThen(t *testing.T) {
	var second bool
	var first bool
	a := promise.New(func() (int, error) {
		return 42, nil
	})
	a.Then(func(x int, err error) (int, error) {
		if x != 42 {
			t.Errorf("First then doesn't equal 42")
		}
		first = true
		return 0, nil
	})
	time.Sleep(100 * time.Millisecond)
	a.Then(func(x int, err error) (int, error) {
		if x != 42 {
			t.Errorf("Second then does not equal 42")
		}
		second = true
		return 0, nil
	})
	time.Sleep(100 * time.Millisecond)
	if !first || !second {
		t.Errorf("First or second didn't get called %t %t", first, second)
	}
}

func TestRace(t *testing.T) {
	a := promise.Resolve(42)
	b := promise.New(func() (int, error) {
		time.Sleep(300 * time.Millisecond)
		return 55, nil
	})
	time.Sleep(100 * time.Millisecond)
	v, _ := promise.Race(a, b).Await()
	if v != 42 {
		t.Errorf("Race finished %d", v)
	}
}
func TestAll(t *testing.T) {
	a := promise.New(func() (int, error) {
		time.Sleep(300 * time.Millisecond)
		return 42, nil
	})
	b := promise.New(func() (int, error) {
		time.Sleep(100 * time.Millisecond)
		return 55, nil
	})
	vals, err := promise.All(a, b).Await()
	if err != nil {
		t.Fatal(err)
	}
	if vals[0] != 42 || vals[1] != 55 {
		t.Errorf("Vals is out of order: %v", vals)
	}

	c := promise.Reject[int](errors.New("blerg"))

	_, err = promise.All(a, b, c).Await()
	if err.Error() != "blerg" {
		t.Errorf("Failed function didn't result in error")
	}
}

func TestTime(t *testing.T) {

	promise.Race(promise.Timeout[int](time.Second), promise.New(func() (int, error) {
		return 42, nil
	})).Await()

}

func TestChannel(t *testing.T) {
	p := promise.Timeout[int](time.Millisecond * 500)
	ch := time.Tick(time.Millisecond * 700)
	counter := int32(5)
	now := time.Now()
	for i := int32(0); i < counter; i++ {
		go func() {
			<-p.Done()
			if time.Since(now) < time.Millisecond*500 {
				t.Error("p.Done() resolved too quickly")
			}
			atomic.AddInt32(&counter, -1)

		}()
	}
	<-ch
	x := atomic.LoadInt32(&counter)
	if x != 0 {
		t.Errorf("Not all go routines returned after selecting on promise.Done(): %d", x)
	}
	ch = time.Tick(time.Millisecond * 100) // Stop us hanging
	select {
	case <-ch:
		t.Error("Timeout waiting for closed channel to signal")
	case <-p.Done():
		return
	}
}

func TestString(t *testing.T) {
	p := promise.Resolve(int32(42))
	str := p.String()
	if str != "Promise.int32" {
		t.Errorf("String is incorrect: %s", str)
	}
}
