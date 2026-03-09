package hashmap

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

// TestLimitLruMap_BasicOps 测试基础操作（Put/Get/Del/Len/Contains/Clear）
func TestLimitLruMap_BasicOps(t *testing.T) {
	lru := NewLimitLruMap[string, int](10)

	// 1. 测试Put和Get
	t.Run("PutAndGet", func(t *testing.T) {
		oldVal, ok := lru.Put("key1", 100)
		if ok || oldVal != 0 {
			t.Errorf("Put新key时，oldVal应返回0，ok应返回false，实际oldVal=%d, ok=%v", oldVal, ok)
		}

		val, ok := lru.Get("key1")
		if !ok || val != 100 {
			t.Errorf("Get key1失败，期望100，实际val=%d, ok=%v", val, ok)
		}

		// 测试更新已存在的key
		oldVal, ok = lru.Put("key1", 200)
		if !ok || oldVal != 100 {
			t.Errorf("Put更新key时，oldVal应返回100，ok应返回true，实际oldVal=%d, ok=%v", oldVal, ok)
		}
		val, ok = lru.Get("key1")
		if !ok || val != 200 {
			t.Errorf("Get更新后的key1失败，期望200，实际val=%d, ok=%v", val, ok)
		}
	})

	// 2. 测试Len和Contains
	t.Run("LenAndContains", func(t *testing.T) {
		if lru.Len() != 1 {
			t.Errorf("Len期望1，实际%d", lru.Len())
		}
		if !lru.Contains("key1") {
			t.Error("Contains(key1)应返回true")
		}
		if lru.Contains("key2") {
			t.Error("Contains(key2)应返回false")
		}
	})

	// 3. 测试Del
	t.Run("Del", func(t *testing.T) {
		lru.Del("key1")
		if lru.Len() != 0 {
			t.Errorf("Del后Len期望0，实际%d", lru.Len())
		}
		if lru.Contains("key1") {
			t.Error("Del后Contains(key1)应返回false")
		}
		// 测试删除不存在的key（无panic）
		lru.Del("key2")
	})

	// 4. 测试Clear
	t.Run("Clear", func(t *testing.T) {
		lru.Put("key1", 1)
		lru.Put("key2", 2)
		lru.Clear()
		if lru.Len() != 0 {
			t.Errorf("Clear后Len期望0，实际%d", lru.Len())
		}
		if lru.Contains("key1") || lru.Contains("key2") {
			t.Error("Clear后不应包含任何key")
		}
	})

	// 5. 测试Range遍历
	t.Run("Range", func(t *testing.T) {
		lru := NewLimitLruMap[string, int](5)
		lru.Put("a", 1)
		lru.Put("b", 2)
		lru.Put("c", 3)

		visited := make(map[string]int)
		lru.Range(func(key string, value int) bool {
			visited[key] = value
			return true // 继续遍历
		})

		if len(visited) != 3 || visited["a"] != 1 || visited["b"] != 2 || visited["c"] != 3 {
			t.Errorf("Range遍历结果错误，visited=%v", visited)
		}

		// 测试中途停止遍历
		count := 0
		lru.Range(func(key string, value int) bool {
			count++
			return count < 2 // 只遍历前2个元素
		})
		if count != 2 {
			t.Errorf("Range中途停止失败，期望遍历2个元素，实际%d", count)
		}
	})

	// 6. 测试不同key类型
	t.Run("DifferentKeyTypes", func(t *testing.T) {
		// int key
		intLru := NewLimitLruMap[int, string](2)
		intLru.Put(123, "test")
		val, ok := intLru.Get(123)
		if !ok || val != "test" {
			t.Errorf("int key测试失败，val=%s, ok=%v", val, ok)
		}

		// float64 key
		floatLru := NewLimitLruMap[float64, string](2)
		floatLru.Put(3.14, "pi")
		val, ok = floatLru.Get(3.14)
		if !ok || val != "pi" {
			t.Errorf("float64 key测试失败，val=%s, ok=%v", val, ok)
		}
	})
}

// TestLimitLruMap_Concurrent 测试并发读写安全性
func TestLimitLruMap_Concurrent(t *testing.T) {
	capacity := 1 << 8
	lru := NewLimitLruMap[int, int](capacity)
	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	// 10个写协程
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := rand.Intn(2000) // 随机key（部分重复）
				lru.Put(key, workerID*1000+j)
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
				lru.Get(key)
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
				lru.Del(key)
			}
		}()
	}

	wg.Wait()

	// 验证最终长度在合理范围（避免并发panic，且长度不为负数）
	length := lru.Len()
	if length < 0 || length > capacity {
		t.Errorf("并发操作后长度异常，期望0~%d，实际%d", capacity, length)
	}
}

// ========== 性能基准测试 ==========
// BenchmarkLimitLruMap_Put 测试Put性能
func BenchmarkLimitLruMap_Put(b *testing.B) {
	lru := NewLimitLruMap[string, int](100000)
	keys := make([]string, b.N)
	// 预生成key，避免基准测试中生成key的性能干扰
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key_%d", i)
	}

	b.ResetTimer() // 重置计时器，排除预生成key的耗时
	for i := 0; i < b.N; i++ {
		lru.Put(keys[i], i)
	}
}

// BenchmarkLimitLruMap_Get 测试Get性能（命中/未命中）
func BenchmarkLimitLruMap_Get(b *testing.B) {
	lru := NewLimitLruMap[string, int](100000)
	keys := genStringKeys(benchSize)
	missKeys := make([]string, benchSize)
	for i := 0; i < benchSize; i++ {
		lru.Put(keys[i], i)
		missKeys = append(missKeys, keys[i]+"_"+strconv.Itoa(i))
	}

	// 测试命中场景
	b.Run("Hit", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = lru.Get(keys[i%benchSize])
		}
	})

	// 测试未命中场景
	b.Run("Miss", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = lru.Get(missKeys[i%benchSize])
		}
	})
}

// BenchmarkLimitLruMap_Concurrent 测试并发读写性能
func BenchmarkLimitLruMap_Concurrent(b *testing.B) {
	lru := NewLimitLruMap[int, int](100000)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 3 {
			case 0: // Put
				lru.Put(i, i)
			case 1: // Get
				_, _ = lru.Get(i)
			case 2: // Del
				lru.Del(i)
			}
			i++
		}
	})
}

// BenchmarkLimitLruMap_Range 测试Range遍历性能
func BenchmarkLimitLruMap_Range(b *testing.B) {
	lru := NewLimitLruMap[int, int](10_0000)
	// 预插入10000个元素
	for i := 0; i < 10_0000; i++ {
		lru.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.Range(func(key int, value int) bool {
			return true
		})
	}
}

func TestGetWithLruRLockRace(t *testing.T) {
	lru := NewLimitLruMap[string, int](1000)
	// 初始化 100 个 key
	for i := 0; i < 100; i++ {
		lru.Put(fmt.Sprintf("key-%d", i), i)
	}

	// 启动 100 个 goroutine 并发读同一个热点 key
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				lru.Get("key-50") // 热点 key
			}
		}()
	}

	// 启动 10 个 goroutine 并发写
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				lru.Put(fmt.Sprintf("new-key-%d", j), j)
			}
		}()
	}

	wg.Wait()
}
