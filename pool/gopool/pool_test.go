// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/pool/gopool
package gopool

import (
	"fmt"
	"testing"
)

func Benchmark_gopool(b *testing.B) {
	b.StopTimer()
	pool := NewPool(100, 100)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for k := 0; k < 10000; k++ {
			pool.Go(func() {
				s := ""
				for i := 0; i < 10; i++ {
					s = fmt.Sprint(s, i)
				}
			})
		}
	}
	fmt.Println(">>>", pool.NumUnExecu())
}

func Benchmark_gof(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for k := 0; k < 10000; k++ {
			go func() {
				s := ""
				for i := 0; i < 10; i++ {
					s = fmt.Sprint(s, i)
				}
			}()
		}
	}
}

func BenchmarkParallel_gopool(b *testing.B) {
	pool := NewPool(100, 100)
	i := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i++
			pool.Go(func() {
				s := ""
				for i := 0; i < 10; i++ {
					s = fmt.Sprint(s, i)
				}
			})
			if i == 3000 {
				// pool.Close()
			}
		}
	})
	fmt.Println(">>>", pool.NumUnExecu())
}

func BenchmarkParallel_gof(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			go func() {
				s := ""
				for i := 0; i < 10; i++ {
					s = fmt.Sprint(s, i)
				}
			}()
		}
	})
}
