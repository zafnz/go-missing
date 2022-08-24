# Timeout implementation

This library includes `TimeoutFn` and `TimeoutFnErr`, which are very useful functions to allow you to timeout a function that doesn't offer such functionality. However this is a problematic area for a number of reasons, so please read this carefully before using this functionality (or anything similar) in production!

If you want the TL;DR for this, scroll down to "Simplifed Usage" section.

## Do it some other way

It is _always_ better to use existing timeout functionality. For http requests, use either the `http.Client`'s timeouts, or the `context.Context` supplied to `http.NewRequestWithTimeout` (you should always use `NewRequestWithTimeout` in prodcution code, never be lazy and use `http.Get`!).

The `context.Context` is very powerful, and a lot of good libraries that have blocking functionality (eg network calls, disk IO, etc) will take a context.Context. If you want to be able to timeout, then use that. 

```
ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
defer cancel() // You should call this as soon as you've finished using this context, eg, the below call finishes.
somelibrary.ThatTakesACtx(ctx, ...)
```

That _should_, if they've written their code right, give you the timeout you are wanting, and it's probably the best way to implement it. (The exception being if there are alternatives they recommend, e.g. gocb prefers it's Timeout value in its options struct).

## Go routines are cooperative

The next most important thing to recognise is that there is no way to terminate a go routine. If a thread is taking a long time, you can't forcefully terminate it. 

That means if you use this code, then your function will keep running, even after "timing out" until it eventually finishes (if ever!). There is no way to forcefully terminate a go routine!

## Go routines share memory 

Because memory is shared, a long running routine that you have long since stopped caring about could alter memory in ways that cause problems later. In fact this is why these `TimeoutFn` and `TimeoutFnErr` functions exist in the first place -- to reduce the problems associated with that.

If you read this and say "use channels", that's what these functions do. They exist to abstract away the uglyness of having to use channels purely for timing out support.

# Example of the problem 

In this example we wish to call some function that will return a string, eventually. But we don't want to wait more than 1 second for the function to return, if it doesn't return in 1 seconds, then we'd rather use a default value and move on. 

```
// We will capture the `str` variable in our timeout, and use that to return...
var str
_, err := TimeoutFn(time.Second * 1, func() (int) {
    // Our function takes too long, so the timeout elapses...
    time.Sleep(2 * time.Second) 
    str = "Set by the function"
    return 0 
})

// Because of the timings above, this will happen...
if err == os.ErrDeadlineExceeded {
    str = "Timeout occured!"
}
// So now str will correctly be set to "Timeout occured"

// However, if we now wait a bit....
time.Sleep(5 * time.Second)

// A few seconds later, str will equal "Set by function", even though we thought a timeout occured.

fmt.Println(str) // Output: Set by the function
```

This can be very problematic if you depend on the variable later on. 

# How to solve it 

In this example we use `TimeoutFn` for it's intended purpose and return the value, as opposed to capturing it. 

As in the previous example the timeout will occur, but unlike the previous example the value of str won't be trashed by the function later on. 
```
str, err := TimeoutFn(time.Second * 1, func() (string) {
    // Our function takes too long, so the timeout elapses...
    time.Sleep(2 * time.Second) 
    someInternalValue := "Set by the function"
    return someInternalValue 
})

// Because of the timings above, this will happen...
if err == os.ErrDeadlineExceeded {
    str = "Timeout occured!"
}
// So now str will correctly be set to "Timeout occured"

// However, if we now wait a bit....
time.Sleep(5 * time.Second)

// This time it works, because the TimeoutFn hasn't captured str as a closure.
fmt.Println(str) // Output: Timeout occured
```

# Simplified usage 

Now to demonstrate the usage simply, we have a function we have to call that returns a struct, but might take too long to do so...
```
maxDuration := time.Second * 2
val, err := TimeoutFn(maxDuration, func() DataStruct {
    val := AnnoyinglyLongFunction()
    return val
})

if err != nil {
    return err 
}
DoSomethingWithData(val)
```

## Returning errors

A common calling syntax is to return a value or error. `TimeoutFnErr` allows you to return an arbitrary value and an error:
```

val, err := TimeoutFnErr(maxDuration, func() DataStruct {
    val, err := AnnoyinglyLongFunctionThatReturnsAnError()
    return val, err // the value of error is returned if this doesn't timeout 
})

if err == os.ErrDeadlineExceeded {
    log.Printf("A timeout occured!")
    return err 
} else if err != nil {
    log.Printf("Some other error occured: %s", err)
    return err 
}
DoSomethingWithData(val)
```

# Final remarks

The TimeoutFn is a powerful tool, but it is important to remember that if a timeout occurs, that go routine will still remain running until it finally (if ever) finishes. 

Remember: **Don't communicate by sharing memory; share memory by communicating.**