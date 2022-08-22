package missing

// Treat slices as objects with methods.

// An AnyList is a slice that can contain anything, but lacks the Contains function (as the type doesn't
// have to be comparable). Compare to a List, which can contain any comparable item, and thus has the Contains
// function
type AnyList[T any] []T

// Append provided values to the end of the list
//  l := List[int]{1,2,3}
//  l.Append(4,5,6)
//  fmt.Printf(l) // [1 2 3 4 5 6]
//
// You can also append a list by using ...
//  l.Append(anotherSlice/List...)
func (l *AnyList[T]) Append(vals ...T) {
	*l = append(*l, vals...)
}

// Same as Append, except prepend the values to the front of the list.
func (l *AnyList[T]) Prepend(vals ...T) {
	*l = append(vals, *l...)
}

// Entirely identical to len(list)
func (l AnyList[T]) Len() int {
	return len(l)
}

// Calls the provided function for each item in the list.
func (l AnyList[T]) Foreach(fn func(T)) {
	for _, v := range l {
		fn(v)
	}
}

func (l AnyList[T]) Reduce(fn func(T, any) any, initial any) any {
	a := initial
	for _, v := range l {
		a = fn(v, a)
	}
	return a
}
