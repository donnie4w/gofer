package hashmap

import (
	"testing"
	"time"
)

func BenchmarkParallelLimitMap(b *testing.B) {
	lm := NewLimitMap[int64, int64](1 << 20)
	b.ResetTimer()
	i := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i++
			k := time.Now().UnixNano() + int64(i)
			lm.Put(k, time.Now().UnixNano())
			if k%5 == 0 {
				lm.Remove(k)
			}
			lm.Get(k)
		}
	})
}

func BenchmarkSerialLimitMap(b *testing.B) {
	lm := NewLimitMap[int64, int64](1 << 17)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := time.Now().UnixNano() + int64(i)
		lm.Put(k, time.Now().UnixNano())
		if k%5 == 0 {
			lm.Remove(k)
		}
		lm.Get(k)
	}
}
