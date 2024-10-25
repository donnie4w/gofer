// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hashmap

import (
	"fmt"
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
				lm.Del(k)
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
			lm.Del(k)
		}
		lm.Get(k)
	}
}

func TestLimitMap(t *testing.T) {
	lm := NewLimitMap[int64, int64](1 << 10)
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
