# Promises for go 
[![GoDoc](https://godoc.org/github.com/zafnz/go-missing/promise?status.svg)](https://godoc.org/github.com/zafnz/go-missing/promise)

Before you send me hatemail, yes, promises are not the way to do things in go. However sometimes it's just easier to use familiar tools, even when they aren't the best way to do things. This library exists to make promise like functionality in a way that operates in a go like fashion. This isn't a complete promise library, functionality that makes no sense in go (like Catch) doesn't exist. That said, if you want to implement it; pull requests welcome!

# Usage 

Simple implementation:
```
p := promise.New(func() (int, error) {
    // Something that might take awhile...
    return 42, nil
})

// Await pauses execution
ultimateAnswer, _ := p.Await()
fmt.Println(ultimateAnswer) // Output: 42
```

Errors:
```
p := promise.New(func() (bool, error) {
    // Calculate if entropy can be reversed, returns a bool pointer or nil if no answer available...
    reversal = multivax.CanEntropyBeReversed()
    if reversal == nil {
        return errors.New("Insufficient data for a meaningful answer.")
    }
    return *reversal
})

// Await pauses execution
lastQuestionAnswer, err := p.Await()
if err != nil {
    fmt.Println(err.Error())
} else {
    // Sadly this won't output.
    fmt.Println(lastQuestionAnswer)
}
// Output: "Insufficient data for a meaningful answer."
```

# Other promise like functionality 

This library has all the usual suspects. The [GoDoc](https://godoc.org/github.com/zafnz/go-missing/promise) has the complete documentation, but the following functionality is available:

- `All(...)` returns a promise that resolves once all promises have been resolved, or any have errored.
- `Race(...)` returns a promise that resolves once any promise has resolved, or any error.
- `Reject(error)` returns a promise that always errors with the provided error.
- `Resolve(any)` returns a promise that resolves immediately with the provided value.
- `Timeout(time.Duration)` returns a promise that will error with `os.ErrDeadlineExceeded` after the specified duration (useful with promise.Race)

As well as each promise offers the following:
- `val, err := promise.Await()` returns the result of the promise or error once the promise has resolved.
- `p := promise.Then(fn)` returns a new promise that will run once the first promise resolves (See section below)

# Then 

It's relatively common in promise implementations to have a Then() functionality, where after a promise resolves specified code executes. That's not quite the go way, and it would typically make more sense to simply call that code from the original promise. However .Then() does offer the ability for multiple go routines to execute once a common result is achieved. Eg perhaps there are multiple go routines that need data from a single promise, then passing those go routines the promise, and them attaching a Then() allows them all to get the value.

## Simple Then() example 
```
finalPromise := promise.New(func() (string, error) {
    return "Hello", nil
}).Then(func(str string, err error) (string, error) {
    return str + " World", nil
}).Then(func(str string, err error) (string, error) {
    return str + "!", nil
})

result, _ := finalPromise.Await()
fmt.Println(result) 
```

## More realistic Then() example:
Here we are fetching some data, presumably it takes awhile to get, but while that's happening we want to do other tasks in parallel, perhaps spawn some worker threads to do other things that will eventually want that data...
```
func worker(data SomeDataType, err error) (SomeDataType, error) {
    // If fetching the big data returned an err, then err will be set. 
    if err != nil {
        return nil, err 
    }
    // Do things with the data 
    DoStuff(data)
    return nil, nil // A Then must return the same type as the original promise, but this worker doesn't need to do that.
}
func beginWork() {
    // Create a promise to retrieve the data...
    bigDataPromise := promise.New(func() (SomeDataType, error) {
        data, err := FetchSomeBigData()
        return data, err 
    })

    // Spawn some worker threads that will use the data once it arrives...
    for i := 0; i < 10; i++ {
        bigDataPromise.Then(worker)
    }
    // other things happen...
}
```

# Promises and multiple workers
Here we have a complex setup, where we have a data structure that takes awhile to retrieve, and have a bunch of workers that need to do other things for awhile, and then process the data once it's available, as well as a special function that needs to run after the data is available.
```
func worker(p promise.Promise[SomeDataType]) error {
    // Do things before the data comes in...
    time.Sleep(rand.Intn(5) * time.Second)
    // Ready for the data now, maybe it's already here, maybe it's not 
    data, err := p.Await()
    if err != nil {
        return err 
    }
    // Now we can use the data!
    err = nowDoThingsWithData(data)
    return err 
}

func beginWork() error {
    // Create a promise to retrieve the data...
    bigDataPromise := promise.New(func() (SomeDataType, error) {
        // This might take awhile...
        data, err := FetchSomeBigData()
        return data, err 
    })

    // Create a bunch of promises that will work on the data...
    var workers []promise.Promise[int]
    for i := 0; i < 10; i++ {
        p := promise.New(func() (int, error) {
            err := worker(bigDataPromise)
            return 0, err // our workers don't return anything except to indicate an error 
        })
        workers = append(workers, p)
    }

    // We also need to process the data in some other way too...
    // A then promise must return the same data type as the original promise (which allows for chaining)
    // but it doesn't _need_ to...
    specialProcess := bigDataPromise.Then(func(data SomeDataType, err error) (SomeDataType, error) {
        if err != nil {
            return err // Retrieving the big data returned an error...
        }
        err = DoOtherThingsWithData(data)
        return data, err // We have to return something, which allows for chaining, so return the original data
    })

    // At this point we could simply return, and let the workers run, but lets say we need to wait for 
    // the workers to finish.

    _, err := promise.All(workers...).Await()
    // If any of the workers returned an err, then err will be set. 
    if err != nil {
        fmt.Printf("A worker returned an error: %s\n", err)
    }
    fmt.Println("All workers have finished handling the data")

    // Maybe we need to wait for our .Then() to finish too...
    _, err = specialProcess.Await()
    if err != nil {
        fmt.Printf("The special handler had an error: %s\n", err)
    }
    fmt.Println("The special handler also finished")
}
```

And there you have it. We've implemented a worker pool, with a WaitGroup, and passed data asynchronisely to each worker, and we even could have used the worker promises to return the results from the workers!
