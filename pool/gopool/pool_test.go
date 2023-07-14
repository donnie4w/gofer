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
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Go(func() {
				s := ""
				for i := 0; i < 10; i++ {
					s = fmt.Sprint(s, i)
				}
			})
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
