// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"fmt"
	"strconv"
	"testing"
)

func Test_bf(t *testing.T) {
	filter := NewBloomFilter(1<<23, 0.01)

	for i := range 31 {
		filter.Add([]byte("apple==>" + strconv.Itoa(i)))
	}

	for i := range 31 {
		bs := []byte("apple==>" + strconv.Itoa(i))
		if !filter.Contains(bs) {
			fmt.Printf("Item '%s' is definitely not present.\n", string(bs))
		}
	}
}

func BenchmarkBloom(b *testing.B) {
	filter := NewBloomFilter(1<<20, 0.001)
	fmt.Println(filter)
	for i := 0; i <= 1<<17; i++ {
		filter.Add([]byte("apple==>" + strconv.Itoa(i)))
	}
	b.ResetTimer()
	i := 0
	b.RunParallel(func(pb *testing.PB) {
		i++
		for pb.Next() {
			s := "apple==>" + strconv.Itoa(i)
			if !filter.Contains([]byte(s)) {
				panic(fmt.Sprintf("Item '%s' is definitely not present.\n", s))
			}
		}
	})
}
