package missing_test

import (
	"testing"
	"time"

	"github.com/zafnz/go-missing"
)

func TestTimeoutFn(t *testing.T) {
	a := 42
	b := 69

	c, err := missing.TimeoutFn(time.Second*1, func() int {
		return a + b
	})
	if err != nil {
		t.Error("Got timeout when not expected.")
	}
	if c != 42+69 {
		t.Error("Didn't sum it up correctly")
	}

	_, err = missing.TimeoutFn(time.Millisecond*100, func() int {
		time.Sleep(time.Millisecond * 500)
		return a + b
	})
	if err == nil {
		t.Error("Did not timeout in 100ms")
	}
	val := 10
	time.Sleep(time.Millisecond * 1000)
	if val != 10 {
		t.Error("Somehow val was modified")
	}

}
