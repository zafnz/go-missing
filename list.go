package missing

// Treat slices as objects with methods.

// A List can contain any comparable type (See `AnyList`` for lists that support any type) and has some useful
// functions associated with it.
type List[T comparable] []T

// Append provided values to the end of the list
//   l := List[int]{1,2,3}
//   l.Append(4,5,6)
//   fmt.Printf(l) // [1 2 3 4 5 6]
//
// You can also append a list by using ...
//    l.Append(anotherSlice/List...)
func (l *List[T]) Append(vals ...T) {
	*l = append(*l, vals...)
}

// Same as Append, except prepend the values to the front of the list.
func (l *List[T]) Prepend(vals ...T) {
	*l = append(vals, *l...)
}

// Inserts the supplied values into the list at the specified index.
func (l *List[T]) Insert(index int, vals ...T) {
	before := (*l)[0:index]
	after := (*l)[index:]
	*l = append(before, vals...)
	*l = append(*l, after...)
}

// Entirely identical to len(list)
func (list List[T]) Len() int {
	return len(list)
}

// Returns true if the list contains val. Note this is a O(n) search. Look at Sets for faster ways.
func (l List[T]) Contains(val T) bool {
	for _, v := range l {
		if v == val {
			return true
		}
	}
	return false
}

// Calls the provided function for each item in the list. Do not modify the list inside the callback
//
// Note: This is another function that is better handled with go's `for _, v := range list`
func (l List[T]) Foreach(fn func(T)) {
	for _, v := range l {
		fn(v)
	}
}

// A really complicated way to do a Reduce on a list. If you insist you can use it, but the go way is clearer
//
// Example if you insist on doing this:
//    list := List[int]{1,2,3,4,5}
//    sum := list.Reduce(func(v int, total any) any {
//        return v + total.(int)
//    }, 100).(int)
//
// Much clearer way:
//    list := List[int]{1,2,3,4,5}
//    sum := 100
//    for _, v := range list {
//        sum += v
//    }
func (l List[T]) Reduce(fn func(T, any) any, initial any) any {
	a := initial
	for _, v := range l {
		a = fn(v, a)
	}
	return a
}
