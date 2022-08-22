package missing

import (
	"os"
	"time"
)

// Generic functions that are missing in go, and clearly needed.

// Returns true if the provided slice (which can be any kind of slice, array, missing.List, etc) contains
// the provided value. Is an O(n) search. (You may find the missing.List or missing.Set a better choice)
func InSlice[T comparable](slice []T, find T) bool {
	for _, val := range slice {
		if val == find {
			return true
		}
	}
	return false
}

// Ternary if. Returns the trueVal if the comparison value is true, otherwise the returns the falseVal.
// Why? Because it's allows inline if statements. Whether this is a good idea is entirely up to you
//
// Example:
//   something := 42
//   fmt.Printf("Something %s divisible by 7\n", missing.If(42 % 7 == 0, "is", "is not"))
//   Output: Something is divisible by 7
func If[T any](cmp bool, trueVal T, falseVal T) T {
	if cmp {
		return trueVal
	} else {
		return falseVal
	}
}

// Calls the supplied function, and returns it's return value, or returns the unitialized value and
// os.ErrDeadlineExceeded if the timeout duration is exceeded. This allows you to call functions with a timeout,
// without having to worry about the implementation details of using a goroutine yourself. So long as you don't
// modify any variables you close over (eg modify any variables from the calling function) then this is probably safe.
//
// Important Note: Go has no way to terminate a goroutine. If your function does not exit, it will remain using
// a go routine thread forever. This is a limitation (or a feature) of go. If you want to support some kind of
// termination ability, then use context.Contexts and signal with those! Almost all built-in libraries that
// can take then (eg http, net, etc) support taking a ctx. You probably don't need this function at that point.
//
// Example -- ensure add a+b and ensure it happens within 1 second:
//	 a := 42
//	 b := 69
//	 c, err := TimeoutFn(time.Second*1, func() int {
// 		return a + b
//	 })
//	 if err != nil {
//		fmt.Println("A timeout occured.")
//	 } else {
//		fmt.Println(c)
//	 }
//
// Note, do not close over and modify variables from the parent, this is why this function exists, instead return
// the values safely in your supplied function. Otherwise you will encounter some sadness.
//
// Example of things going horribly wrong
//
//    // We have a function that captures `str` and will set it to "Set by function" after 2 seconds. However
//    // in this example it will timeout after 1 second...
//    var str
//     _, err := TimeoutFn(time.Second * 1, func() (string) {
//           time.Sleep(2 * time.Second)
//           str = "Set by function"
//           return ""
//     })
//
//     // Because of the timings above, this will happen...
//     if err == os.ErrDeadlineExceeded {
//          str = "Timeout occured!"
//     }
//
//     // However, if we now wait a bit....
//     time.Sleep(5 * time.Second)
//
//     // A few seconds later, str will equal "Set by function", even though we thought a timeout occured.
//     fmt.Println(str) // Outputs: Set by function
// Safer way:
//     // Here we instead return str as a value from our timeout protect function.
//     str, err := TimeoutFn(time.Second * 1, func() (string) {
//           time.Sleep(2 * time.Second)
//           return "Set by function"
//     })
//     if err == os.DeadlineExceeded {
//           str = "Timeout!"
//     }
//     // Now, no matter what, if a timeout occurs, str will remain "Timeout!", even if the provided function
//     // eventually returns.
//     time.Sleep(10 * time.Second)
//     fmt.Println(str) // Outputs: "Timeout!"
func TimeoutFn[T any](duration time.Duration, fn func() T) (T, error) {
	ch := make(chan T)
	go func() {
		r := fn()
		ch <- r
	}()
	select {
	case ret := <-ch:
		return ret, nil
	case <-time.After(duration):
		var r T
		return r, os.ErrDeadlineExceeded
	}
}

// Similar to TimeoutFn, but the function you provide returns two values, one of them an error, and TimeoutFnErr
// will return those two values. If the function times out, then err will be os.ErrDeadlineExceeded
//
// See notes for TimeoutFn for important information.
//
// Example (Note: this is a bad example, a http.Request can be created with a context.Context, which is a far better
// better way to do timeouts):
//   resp, err := TimeoutFnErr(time.Second * 10, func() (*http.Response, error) {
//     resp, err := http.Do(req)
//     return resp, err
//   })
func TimeoutFnErr[T any](duration time.Duration, fn func() (T, error)) (T, error) {
	ch := make(chan T)
	errCh := make(chan error)
	go func() {
		val, err := fn()
		ch <- val
		errCh <- err
	}()
	select {
	case val := <-ch:
		err := <-errCh
		return val, err
	case <-time.After(duration):
		var r T
		return r, os.ErrDeadlineExceeded
	}
}
