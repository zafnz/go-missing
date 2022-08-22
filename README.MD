# The missing go functions and types.

Go is a powerful language, but some people believe quite strongly that it is missing some key functionality.
This library attempts to address that. Contributions are welcome. 

# Promises
This library also includes promises, see [http://github.com/zafnz/go-missing/promise](https://github.com/zafnz/go-missing/tree/main/promise)

# Usage (optional alias)
Just like any other library, `go get github.com/zafnz/go-missing`.

While you can use this library like any other, the `missing` prefix for every type and function can be a bit 
annoying. So you might want to do something like. 

```go
import M "github.com/zafnz/missing"

func main() {
    var mylist M.List[int]
    resp, err := M.TimeoutFnErr(time.Second * 5, http.Get("http://example.com/")) 

    fmt.Printf("7 * 3 %s 21", 
        M.If(7 * 3 == 21, "is", "is not"))
}
```

# List methods.
A slice of a comparable type that has some additional methods (see GenericList for any type).

With a normal slice the `x = append(x, val)` is not exactly intuative, and seems quaint when considering other 
languages. It is quite reasonable as to why, but the `missing.List` type makes slices more object-oriented. 
Note: Underneath a `List` is just a slice, and you can do anythign to a `List` that you can to a slice. 

A `missing.List` is a slice that has some useful functions -- see godoc for full documentation, but an overview is: 

- `Append(vals...)` // Appends the values to the list
- `Prepend(vals...)` // Prepends the values to the list
- `Contains(val)` // Returns true if the list contains a value 
- `Insert(idx, vals...)` // Inserts the vals at the specified index in the list. Panics if out of bounds
- `Len()` // Returns the length of the list. This is literally just `len(list)`, but is here because why not?
- `Foreach(fn)` // Calls the provided function for each item in the list. You probably just want to use a normal for loop
- `Sort(fn)` // Sorts the list inplace.  

# AnyList 
An any list is a list, but is missing `Contains` as unlike a `List` a generic list can be any type, it doesn't 
need to be comparable. Otherwise it is functionaly identical to List.
