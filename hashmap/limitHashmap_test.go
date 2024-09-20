package hashmap

import (
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
