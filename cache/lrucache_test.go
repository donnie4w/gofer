// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"testing"
	"time"
)

func BenchmarkParallelLru(b *testing.B) {
	lc := NewLruCache[int64](1 << 20)

	for i := 0; i < 1<<18; i++ {
		k := int64(i)
		lc.Add(k, time.Now().UnixNano())
	}

	b.ResetTimer()
	i := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i++
			k := int64(i)
			//lc.add(k, time.Now().UnixNano())
			//if k%5 == 0 {
			//	lc.Remove(k)
			//}
			lc.Get(k)
		}
	})
}

func BenchmarkSerialLru(b *testing.B) {
	lc := NewLruCache[int64](1 << 20)
	for i := 0; i < 1<<18; i++ {
		k := int64(i)
		lc.Add(k, time.Now().UnixNano())
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := int64(i)
		//lc.add(k, time.Now().UnixNano())
		//if k%5 == 0 {
		//	lc.Remove(k)
		//}
		lc.Get(k)
	}
}
