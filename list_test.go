package missing_test

import (
	"testing"

	"github.com/zafnz/go-missing"
)

func TestList(t *testing.T) {
	x := missing.List[int]{}
	x.Append(1, 2, 3)
	if x[0] != 1 && x[2] != 3 {
		t.Error("List doesn't append items in provided order")
	}
	x.Prepend(0)
	if x[0] != 0 && x[1] != 1 {
		t.Error("List doesn't prepend items in provided order")
	}
}

func TestListInsert(t *testing.T) {
	x := missing.List[int]{1, 2, 3, 4, 10, 11, 12, 13}
	y := []int{5, 6, 7, 8, 9}
	x.Insert(4, y...)
	if x[0] != 1 || x[4] != 5 || x[8] != 9 || x[9] != 10 {
		t.Errorf("Insert did something wrong. List is supposed to be 1...13, but is: %v", x)
	}
}

func TestReduce(t *testing.T) {
	x := missing.List[int]{1, 2, 3, 4, 5}

	sum := x.Reduce(func(i int, a any) any {
		return i + a.(int)
	}, 10)

	if sum.(int) != 25 {
		t.Errorf("Oddly, 1+2+3+4+5+10 does not equal 25: %v", sum)
	}
}
