package missing

import (
	"encoding/json"
	"fmt"
)

// Under the hood a set is simply a map:
//   type Set[T comparable] map[T]struct{}
// All the usual map functions will work (such as len).
//
// Note: You can instantiate a Set with `make(missing.Set[T])`, `missing.NewSet()` or just:
//   set := missing.Set[int]{}
type Set[T comparable] map[T]struct{}

// Creates a new set, using the provided slice.
//    x := []int { 1,2,3,4,5,6 }
//    s := set.NewSet(x)
func NewSet[T comparable](slice []T) Set[T] {
	s := make(Set[T], len(slice))
	for _, v := range slice {
		s[v] = struct{}{}
	}
	return s
}

// Creates a copy of the set as a slice. Useful for giving the set to other functions that expect a slice.
func (s Set[T]) ToSlice() []T {
	keys := make([]T, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return keys
}

// Returns true if the set contains the provided value. Does a direct == comparison. Which is to say if your
// set contains pointers to objects, then it returns true if it is the same pointer. If your set contains
// the objects themselves, then returns true if the objects are equivilent.
func (s Set[T]) Contains(v T) bool {
	_, found := (s)[v]
	return found
}

// Returns the length of a set. (Note: this is functionaly equivilent as `len(s)`)
func (s Set[T]) Length() int {
	return len(s)
}

// Adds the values from the supplied set to this set. (inplace operation, c.f. Union)
func (a *Set[T]) AddSet(b Set[T]) {
	for k := range b {
		(*a)[k] = struct{}{}
	}
}

// Add individual value(s) to the set.
func (s *Set[T]) Add(vals ...T) {
	for _, v := range vals {
		(*s)[v] = struct{}{}
	}
}

// Adds the records from the slice to this set inplace. Functionaly the same as s.Add(slice...)
func (s *Set[T]) AddSlice(slice []T) {
	for _, v := range slice {
		(*s)[v] = struct{}{}
	}
}

// Returns the difference of this set and the provided set. (a - b). That is to say:
//   a := MakeSet([]int{1,2,3,4,5})
//   b := MakeSet([]int{1,2,3})
//   d := a.Difference(b)
// d contains 4,5
func (a Set[T]) Difference(b Set[T]) Set[T] {
	diff := make(Set[T], len(a))
	for v := range a {
		if _, found := (b)[v]; !found {
			diff[v] = struct{}{}
		}
	}
	return diff
}

// Returns the union of set a + b. This is the equivilent of set.AddSet, but returns
// a new set instead of being an inplace operation.
func (a Set[T]) Union(b Set[T]) Set[T] {
	union := make(Set[T])
	for k := range a {
		union[k] = struct{}{}
	}
	for k := range b {
		union[k] = struct{}{}
	}
	return union
}

// Returns the intersection of sets a and b, that is to say only things that are in both 'a'
// and 'b'
func (a Set[T]) Intersection(b Set[T]) Set[T] {
	intersection := make(Set[T], len(a))
	for v := range a {
		if _, found := b[v]; found {
			intersection[v] = struct{}{}
		}
	}
	return intersection
}

// A string representation of the set (essentially returns a string formated list)
func (s Set[T]) String() string {
	return fmt.Sprint(s.ToSlice())
}

// A set marshalls into a json array.
func (s Set[T]) MarshalJSON() ([]byte, error) {
	list := s.ToSlice()
	return json.Marshal(list)
}

// A set unmarshalls from a json array.
func (s *Set[T]) UnmarshalJSON(b []byte) error {
	var list []T
	err := json.Unmarshal(b, &list)
	if err != nil {
		return err
	}
	(*s) = Set[T]{}
	s.Add(list...)
	return nil
}
