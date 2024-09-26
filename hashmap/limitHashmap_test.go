package hashmap

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkParallelLimitHashMap(b *testing.B) {
	lm := NewLimitHashMap[uint64, int64](1 << 17)
	b.ResetTimer()
	var i int64 = 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			k := uint64(time.Now().UnixNano() + atomic.AddInt64(&i, 1))
			lm.Put(k, time.Now().UnixNano())
			if k%9 == 0 {
				lm.Del(k)
			}
			if _, ok := lm.Get(k); !ok {
			}
		}
	})
}

func BenchmarkParallelLimitHashMapGet(b *testing.B) {
	lm := NewLimitHashMap[int, int64](1 << 17)
	for i := range 500000 {
		lm.Put(i, time.Now().UnixNano())
	}
	b.ResetTimer()
	var i int = 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i++
			if _, ok := lm.Get(i); !ok {
			}
		}
	})
}

func BenchmarkSerialLimitHashMap(b *testing.B) {
	lm := NewLimitHashMap[uint64, int64](1 << 17)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := uint64(time.Now().UnixNano() + int64(i))
		lm.Put(k, time.Now().UnixNano())
		if k%9 == 0 {
			lm.Del(k)
		}
		lm.Get(k)
	}
}

func TestLimitHashMap(t *testing.T) {
	lm := NewLimitHashMap[int64, int64](1 << 10)
	for i := 0; i < 1000; i++ {
		k := time.Now().UnixNano() + int64(i)
		lm.Put(k, time.Now().UnixNano())
		if k%3 == 0 {
			lm.Del(k)
		}
		lm.Get(k)
	}
	fmt.Println(lm.Len())
}
