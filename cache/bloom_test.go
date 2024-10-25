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

func BenchmarkParallelBloom(b *testing.B) {
	bf := NewBloomFilter(1 << 20)
	b.ResetTimer()
	i := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i++
			k := []byte(strconv.FormatInt(int64(i), 10))
			bf.Add(k)
			bf.Contains(k)
		}
	})
}

func BenchmarkParallelBloomGet(b *testing.B) {
	bf := NewBloomFilter(1 << 20)
	for i := range 1 << 18 {
		k := []byte(strconv.FormatInt(int64(i), 10))
		bf.Add(k)
	}
	b.ResetTimer()
	i := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i++
			k := []byte(strconv.FormatInt(int64(i), 10))
			bf.Contains(k)
		}
	})
}

func BenchmarkSerialBloom(b *testing.B) {
	bf := NewBloomFilter(1 << 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		i++
		k := []byte(strconv.FormatInt(int64(i), 10))
		bf.Add(k)
		bf.Contains(k)
	}
}

func TestBloom(t *testing.T) {
	bf := NewBloomFilter(1 << 10)
	for i := 0; i < 2030; i++ {
		k := []byte(strconv.FormatInt(int64(i), 10))
		bf.Add(k)
	}
	for i := 0; i < 2030; i++ {
		k := []byte(strconv.FormatInt(int64(i), 10))
		if !bf.Contains(k) {
			fmt.Println(string(k))
		}
	}
}
