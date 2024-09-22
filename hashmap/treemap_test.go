// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/gofer/hashmap

package hashmap

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkParalleltreeMap(b *testing.B) {
	tm := NewTreeMap[uint64, int64](64)
	b.ResetTimer()
	var i int64 = 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			k := uint64(time.Now().UnixNano() + atomic.AddInt64(&i, 1))
			tm.Put(k, time.Now().UnixNano())
			if k%9 == 0 {
				tm.Del(k)
			}
			if _, ok := tm.Get(k); !ok {
			}
		}
	})
}

func BenchmarkSerialLimittreeMap(b *testing.B) {
	tm := NewTreeMap[uint64, int64](32)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := uint64(time.Now().UnixNano() + int64(i))
		tm.Put(k, time.Now().UnixNano())
		if k%9 == 0 {
			tm.Del(k)
		}
		tm.Get(k)
	}
}

func Test_treeMap(t *testing.T) {
	tm := NewTreeMap[int, int64](32)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		go func(i int) {
			wg.Add(1)
			defer wg.Done()
			tm.Put(i, time.Now().UnixNano())
		}(i)
	}
	wg.Wait()
	tm.Ascend(func(i int, i2 int64) bool {
		fmt.Println(i, ":", i2)
		return true
	})

	tm.Descend(func(i int, i2 int64) bool {
		fmt.Println(i, ":", i2)
		return true
	})

	tm.Put(10000, 10000)
	fmt.Println(tm.Put(10000, 10001))
	fmt.Println(tm.Put(10000, 10002))
}
