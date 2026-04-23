package hashmap

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

// ========== 正确性测试 ==========

// TestLimitFifoMap_BasicOps 测试基础操作（Put/Get/Del/Len/Contains/Clear）
func TestLimitFifoMap_BasicOps(t *testing.T) {
	fifo := NewLimitFifoMap[string, int](10)

	// 1. 测试Put和Get
	t.Run("PutAndGet", func(t *testing.T) {
		oldVal, ok := fifo.Put("key1", 100)
		if ok || oldVal != 0 {
			t.Errorf("Put新key时，oldVal应返回0，ok应返回false，实际oldVal=%d, ok=%v", oldVal, ok)
		}

		val, ok := fifo.Get("key1")
		if !ok || val != 100 {
			t.Errorf("Get key1失败，期望100，实际val=%d, ok=%v", val, ok)
		}

		// 测试更新已存在的key
		oldVal, ok = fifo.Put("key1", 200)
		if !ok || oldVal != 100 {
			t.Errorf("Put更新key时，oldVal应返回100，ok应返回true，实际oldVal=%d, ok=%v", oldVal, ok)
		}
		val, ok = fifo.Get("key1")
		if !ok || val != 200 {
			t.Errorf("Get更新后的key1失败，期望200，实际val=%d, ok=%v", val, ok)
		}
	})

	// 2. 测试Len和Contains
	t.Run("LenAndContains", func(t *testing.T) {
		if fifo.Len() != 1 {
			t.Errorf("Len期望1，实际%d", fifo.Len())
		}
		if !fifo.Contains("key1") {
			t.Error("Contains(key1)应返回true")
		}
		if fifo.Contains("key2") {
			t.Error("Contains(key2)应返回false")
		}
	})

	// 3. 测试Del
	t.Run("Del", func(t *testing.T) {
		fifo.Del("key1")
		if fifo.Len() != 0 {
			t.Errorf("Del后Len期望0，实际%d", fifo.Len())
		}
		if fifo.Contains("key1") {
			t.Error("Del后Contains(key1)应返回false")
		}
		// 测试删除不存在的key（无panic）
		fifo.Del("key2")
	})

	// 4. 测试Clear
	t.Run("Clear", func(t *testing.T) {
		fifo.Put("key1", 1)
		fifo.Put("key2", 2)
		fifo.Clear()
		if fifo.Len() != 0 {
			t.Errorf("Clear后Len期望0，实际%d", fifo.Len())
		}
		if fifo.Contains("key1") || fifo.Contains("key2") {
			t.Error("Clear后不应包含任何key")
		}
	})

	// 5. 测试不同key类型
	t.Run("DifferentKeyTypes", func(t *testing.T) {
		// int key
		intFifo := NewLimitFifoMap[int, string](2)
		intFifo.Put(123, "test")
		val, ok := intFifo.Get(123)
		if !ok || val != "test" {
			t.Errorf("int key测试失败，val=%s, ok=%v", val, ok)
		}

		// float64 key
		floatFifo := NewLimitFifoMap[float64, string](2)
		floatFifo.Put(3.14, "pi")
		val, ok = floatFifo.Get(3.14)
		if !ok || val != "pi" {
			t.Errorf("float64 key测试失败，val=%s, ok=%v", val, ok)
		}
	})
}

// TestLimitFifoMap_Concurrent 测试并发读写安全性
func TestLimitFifoMap_Concurrent(t *testing.T) {
	capacity := 1 << 8
	fifo := NewLimitFifoMap[int, int](capacity)
	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	// 10个写协程
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := rand.Intn(2000) // 随机key（部分重复）
				fifo.Put(key, workerID*1000+j)
			}
		}(i)
	}

	// 10个读协程
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := rand.Intn(2000)
				fifo.Get(key)
			}
		}()
	}

	// 2个删除协程
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				key := rand.Intn(2000)
				fifo.Del(key)
			}
		}()
	}

	wg.Wait()

	// 验证最终长度在合理范围（避免并发panic，且长度不为负数）
	length := fifo.Len()
	if length < 0 || length > capacity {
		t.Errorf("并发操作后长度异常，期望0~%d，实际%d", capacity, length)
	}
}

// ========== 性能基准测试 ==========

// BenchmarkLimitFifoMap_Put 测试Put性能
func BenchmarkLimitFifoMap_Put(b *testing.B) {
	fifo := NewLimitFifoMap[string, int](benchSize)
	keys := make([]string, benchSize)
	// 预生成key，避免基准测试中生成key的性能干扰
	for i := 0; i < benchSize; i++ {
		keys[i] = fmt.Sprintf("key_%d", i)
	}

	b.ResetTimer() // 重置计时器，排除预生成key的耗时
	for i := 0; i < b.N; i++ {
		fifo.Put(keys[i%benchSize], i)
	}
}

// BenchmarkLimitFifoMap_Get 测试Get性能（命中/未命中）
func BenchmarkLimitFifoMap_Get(b *testing.B) {
	fifo := NewLimitFifoMap[string, int](benchSize)
	keys := genStringKeys(benchSize)
	missKeys := make([]string, benchSize)
	for i := 0; i < benchSize; i++ {
		fifo.Put(keys[i], i)
		missKeys = append(missKeys, keys[i]+"_"+strconv.Itoa(i))
	}

	// 测试命中场景
	b.Run("Hit", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = fifo.Get(keys[i%benchSize])
		}
	})

	// 测试未命中场景
	b.Run("Miss", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = fifo.Get(missKeys[i%benchSize])
		}
	})
}

// BenchmarkLimitFifoMap_Concurrent 测试并发读写性能
func BenchmarkLimitFifoMap_Concurrent(b *testing.B) {
	fifo := NewLimitFifoMap[int, int](100000)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 3 {
			case 0: // Put
				fifo.Put(i, i)
			case 1: // Get
				_, _ = fifo.Get(i)
			case 2: // Del
				fifo.Del(i)
			}
			i++
		}
	})
}

// BenchmarkLimitFifoMap_Range 测试Range遍历性能
func BenchmarkLimitFifoMap_Range(b *testing.B) {
	fifo := NewLimitFifoMap[int, int](10_0000)
	for i := 0; i < 10_0000; i++ {
		fifo.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fifo.Range(func(key int, value int) bool {
			return true
		})
	}
}

func TestGetWithFifoRLockRace(t *testing.T) {
	fifo := NewLimitFifoMap[string, int](1000)
	for i := 0; i < 100; i++ {
		fifo.Put(fmt.Sprintf("key-%d", i), i)
	}

	// 启动 100 个 goroutine 并发读同一个热点 key
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				fifo.Get("key-50") // 热点 key
			}
		}()
	}

	// 启动 10 个 goroutine 并发写
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				fifo.Put(fmt.Sprintf("new-key-%d", j), j)
			}
		}()
	}

	wg.Wait()
}
