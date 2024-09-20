package cache

import (
	"testing"
	"time"
)

func BenchmarkParallelLru(b *testing.B) {
	lc := NewLruCache[int64](1 << 20)
	b.ResetTimer()
	i := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i++
			k := time.Now().UnixNano() + int64(i)
			lc.Add(k, time.Now().UnixNano())
			if k%5 == 0 {
				lc.Remove(k)
			}
			lc.Get(k)
		}
	})
}

func BenchmarkSerialLru(b *testing.B) {
	lc := NewLruCache[int64](1 << 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := time.Now().UnixNano() + int64(i)
		lc.Add(k, time.Now().UnixNano())
		if k%5 == 0 {
			lc.Remove(k)
		}
		lc.Get(k)
	}
}
