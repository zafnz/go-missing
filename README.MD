# The missing go functions and types.
[![GoDoc](https://godoc.org/github.com/zafnz/go-missing?status.svg)](https://godoc.org/github.com/zafnz/go-missing)

Go is a powerful language, but some people believe quite strongly that it is missing some key functionality.
This library attempts to address that. Contributions are welcome. 

# Promises
This library also includes promises, see [http://github.com/zafnz/go-missing/promise](https://github.com/zafnz/go-missing/tree/main/promise)
[![GoDoc](https://godoc.org/github.com/zafnz/go-missing/promise?status.svg)](https://godoc.org/github.com/zafnz/go-missing/promise)

# Usage
Just like any other library, `go get github.com/zafnz/go-missing`.

# Functionality

## If
```
func If[T any](condition bool, valTrue T, valFalse T) T
fmt.Printf("5 %s greater than 7\n", missing.If(5 > 7, "is", "is not"))
```
Provides a ternary or inline if statement. Go is deliberately missing this functionality, but if you want it back, here it is. This
is functionality equivilent to the likes of C's: `printf("5 %is greater than 7\n", 5 > 7 ? "is" : "is not")`

## InSlice
```
func InSlice[T comparable](slice []T, val T) bool
```
Returns true if the provided value is in the provided array/slice. Note: This is an O(N) operation. See `missing.Set` for a better 
implementation of this.

## Timeouts
```
val, err := missing.TimeoutFn[T any](time.Duration, func() (T)) (T, error)
val, err := missing.TimeoutFnErr[T any](time.Duration, func() (T, error)) (T, error)

```
Executes the provided function, and returns the resultant return value, or returns an `os.ErrDeadlineExceeded` error if the supplied timeout 
expires. This allows wrapping a function that doesn't support a `context.Context` with timeout functionality. 

The difference between `TimeoutFn` and `TimeoutFnErr` is that a `TimeoutFnErr` supplied function can return an error, which will be supplied
to the caller.

See the TIMEOUT.MD file for a much deeper exploration of this subject, including some significant gotchas with most golang timeout wrappers.

## List methods
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

## AnyList 
An any list is a list, but is missing `Contains` as unlike a `List` a generic list can be any type, it doesn't 
need to be comparable. Otherwise it is functionaly identical to List.

## Sets
A generic Set type, which is a list of objects (must be a comparable type) that is guarenteed to be unique. Converts back and forth between
slice and list trivially and json encodes/decodes as a json array.

- `x := Set[int]{}` // Creates a new empty set of type int.
- `x := make(Set[string])` // Creates an empty set of type string.
- `x := NewSet(slice)` // Creates a new set from a provided slice, of whatever type the slice is.
- `set.Add(items...)` // Adds one or more items to a set.
- `set.Contains(items)` // Returns true if the set contains the specified item
- `set.Union(b)`, `.Intersection(b)`, `.Difference(b)` // Returns a set that is the result of the specified set operation. (See pkg docs)
- `set.AddSlice(slice)` // Adds a slice into the set
- `set.AddSet(set)` // Adds a set into this set (c.f. `.Union` which returns a new set).

# Alias module
While you can use this library like any other, the `missing` prefix for every type and function can be a bit 
annoying. So you might want to do something like: 

```go
import M "github.com/zafnz/missing"

func main() {
    var mylist M.List[int]
    resp, err := M.TimeoutFnErr(time.Second * 5, http.Get("http://example.com/")) 

    fmt.Printf("7 * 3 %s 21", 
        M.If(7 * 3 == 21, "is", "is not"))
}
```
Note: That's entirely optional
