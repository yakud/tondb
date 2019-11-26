package cache

import (
	"testing"
	"time"
)

type testVal struct {
	A []int64
	B uint32
}

func TestNewWithTimer(t *testing.T) {
	val := testVal{}
	c := NewWithTimer(time.Millisecond * 100)

	time.After(time.Millisecond * 200)
}
