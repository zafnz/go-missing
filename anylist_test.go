package missing_test

import (
	"testing"

	"github.com/zafnz/go-missing"
)

func TestAnyList(t *testing.T) {
	x := missing.AnyList[int]{}
	x.Append(1, 2, 3)
	if x[0] != 1 && x[2] != 3 {
		t.Error("List doesn't append items in provided order")
	}
	x.Prepend(0)
	if x[0] != 0 && x[1] != 1 {
		t.Error("List doesn't prepend items in provided order")
	}
}

func TestAnyReduce(t *testing.T) {
	x := missing.AnyList[int]{1, 2, 3, 4, 5}

	sum := x.Reduce(func(i int, a any) any {
		return i + a.(int)
	}, 10)

	if sum.(int) != 25 {
		t.Errorf("Oddly, 1+2+3+4+5+10 does not equal 25: %v", sum)
	}
}
