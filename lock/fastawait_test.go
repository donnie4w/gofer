package lock

import (
	"testing"
	"time"
)

func TestFastAwait(t *testing.T) {
	aw := NewFastAwait[int]()
	for i := 0; i < 10; i++ {
		idx := int64(i)
		go func(idx int64) {
			time.Sleep(1 * time.Second)
			aw.CloseAndPut(idx, int(idx))
		}(idx)
		v, err := aw.Wait(idx, time.Second)
		t.Log(err, idx, v)
	}
	time.Sleep(3 * time.Second)
}
