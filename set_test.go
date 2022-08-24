package missing_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/zafnz/go-missing"
)

func TestSets(t *testing.T) {
	a := missing.Set[int]{}
	if a == nil {
		t.Fatal("missing.NewSet is nil!")
	}
	if a.Length() != 0 {
		t.Error("set from empty slice is not empty")
	}
	a = missing.NewSet([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	if a.Length() != 10 {
		t.Error("Set 1-10 is not len=10")
	}
	a.Add(11, 12)
	if !a.Contains(11) || !a.Contains(12) {
		t.Error("Set doesn't contain new items")
	}
	if a.Length() != 12 {
		t.Error("Set with 2 extra items didn't grow by 2.")
	}
}

func TestAdds(t *testing.T) {
	a := missing.NewSet([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	a.AddSlice([]int{11, 12})
	if !a.Contains(12) {
		t.Error("a.AddSlice didn't add all entries")
	}
	b := missing.NewSet([]int{20, 30, 40})
	a.AddSet(b)
	if !a.Contains(1) || !a.Contains(40) {
		t.Error("a.AddSet doesn't contain new items, or is now missing old")
	}
}

func TestTypes(t *testing.T) {
	s := missing.Set[string]{}
	s.Add("test")
	if !s.Contains("test") {
		t.Error("Sets with strings not working")
	}
	type custom struct {
		blah  string
		other int
	}
	a := custom{"test1", 1}
	b := custom{"test2", 2}

	c := make(missing.Set[custom])
	c.Add(a)
	c.Add(b)
	if c.Length() != 2 {
		t.Error("Custom type set length wrong")
	}
	c.Add(custom{"test1", 1})
	if c.Length() != 2 {
		t.Error("Added same values for custom type, but set increased")
	}
}

func ExampleSet_Union() {
	a := missing.NewSet([]int{1, 2, 3, 4, 5})
	b := missing.NewSet([]int{1, 2, 3})
	d := a.Difference(b)
	fmt.Println(d)
	// Will output 4, 5
}

func TestUnion(t *testing.T) {
	a := missing.NewSet([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	b := missing.NewSet([]int{6, 7, 8, 9, 10, 20, 30, 40})

	u := a.Union(b)
	if !u.Contains(1) || !u.Contains(2) {
		t.Error("a.Union(b) doesn't contain 'a' items")
	}
	if !u.Contains(20) || !u.Contains(30) {
		t.Error("a.Union(b) doesn't contain 'b' items")
	}
	if a.Contains(30) || b.Contains(1) {
		t.Error("a or b now contains something from b or a")
	}
}
func TestDifference(t *testing.T) {
	a := missing.NewSet([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	b := missing.NewSet([]int{6, 7, 8, 9, 10, 20, 30, 40})
	// Everything in a that's not in b
	d := a.Difference(b)

	if d.Contains(6) || d.Contains(10) {
		t.Error("a.Difference(b) contains elements that are in b as well")
	}
	if d.Contains(20) || d.Contains(40) {
		t.Error("a.Difference(b) contains elements that were only in b")
	}
}

func TestIntersection(t *testing.T) {
	a := missing.NewSet([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	b := missing.NewSet([]int{6, 7, 8, 9, 10, 20, 30, 40})
	// Everything in a that's not in b
	i := a.Intersection(b)

	if i.Contains(1) || i.Contains(5) {
		t.Error("a.Intersection(b) return contains items only in 'a'")
	}
	if i.Contains(20) || i.Contains(40) {
		t.Error("a.Intersection(b) return contains items only in 'b'")
	}
	if !i.Contains(6) || !i.Contains(10) {
		t.Error("a.Intersection(b) return does not contain some common elements")
	}

}

func TestToSlice(t *testing.T) {
	originalSlice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	a := missing.NewSet(originalSlice)

	slice := a.ToSlice()
	if len(slice) != 10 {
		t.Error("10 element set didn't become slice of 10 items")
	}
	// Order isn't preserved, so we can't just do slice[0], lets just iterate through all
	// items. Not exactly efficient, but meh.
	for _, i := range slice {
		found := false
		for _, j := range originalSlice {
			if i == j {
				found = true
			}
		}
		if !found {
			t.Fatalf("Element %d in original set is not in slice from set", i)
		}
	}

	// Test whether we got an actual slice back, or an array. We can extend a slice.
	slice = append(slice, 98, 99, 100)
	if len(slice) != 13 {
		t.Error("Slice didn't extend")
	}
}

func TestJson(t *testing.T) {
	a := missing.NewSet([]int{777, 12345, 898989})
	bytes, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	// The bytes should contain some of our distinctive values
	str := string(bytes)
	if !strings.HasPrefix(str, "[") || !strings.HasSuffix(str, "]") {
		t.Fatalf("Set didn't marshall into an array: %s", str)
	}
	if !strings.Contains(str, "777") || !strings.Contains(str, "12345") || !strings.Contains(str, "898989") {
		t.Fatalf("Marshalling set %v does not contain those values in the output: %s", a, str)
	}

	var b missing.Set[int]
	err = json.Unmarshal(bytes, &b)
	if err != nil {
		t.Fatal(err)
	}
	if b.Length() != a.Length() {
		t.Error("Did not unmarshall into set of same length")
	}
	// This should be zero, since everything in b is in a
	d := a.Difference(b)
	if d.Length() != 0 {
		t.Error("Some things in original set not in unmarshalled version")
	}
}

func TestFormatter(t *testing.T) {
	a := missing.NewSet([]int{123, 456})
	str := fmt.Sprintf("%+v", a)
	// We can't dictate the order.
	if str != "[123 456]" && str != "[456 123]" {
		t.Errorf("Set %+v did not format as a list. Got %s", a, str)
	}
}
func TestToString(t *testing.T) {
	a := missing.NewSet([]int{123, 456})
	str := fmt.Sprint(a)
	if str != "[123 456]" && str != "[456 123]" {
		t.Errorf("Set %+v did not format as a list. Got %s", a, str)
	}
}
