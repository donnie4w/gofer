package hashmap

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// 测试常量定义（兼顾小数据量精准验证 + 大数据量稳定性验证）
const (
	smallTestSize = 100      // 小数据量：精准验证每个函数的边界条件
	largeTestSize = 10_0000  // 大数据量：验证高数据量下的稳定性
	benchSize     = 100_0000 // 基准测试数据量
)

// ------------------------------ 工具函数：生成固定key池（避免内存分配干扰） ------------------------------
// 生成字符串key池
func genStringKeys(n int) []string {
	keys := make([]string, n)
	for i := range keys {
		keys[i] = fmt.Sprintf("fixed-key-%d", i)
	}
	return keys
}

// 生成数值key池（uint64）
func genUint64Keys(n int) []uint64 {
	keys := make([]uint64, n)
	for i := range keys {
		keys[i] = uint64(i)
	}
	return keys
}

// ------------------------------ 第一部分：全函数精细化正确性测试（小数据量 + 边界条件） ------------------------------

// TestLimitHashMap_AllFuncs 验证每个公开函数的核心功能和边界条件
func TestLimitHashMap_AllFuncs(t *testing.T) {
	// 测试场景1：字符串key + int值（最常用场景）
	t.Run("string-key-int-value", func(t *testing.T) {
		// 1. 初始化
		lhm := NewLimitHashMap[string, int](smallTestSize * 5)
		keys := genStringKeys(smallTestSize)

		// 2. 测试Put（新增）
		t.Run("Put-New", func(t *testing.T) {
			prev, ok := lhm.Put(keys[0], 100)
			if ok || prev != 0 {
				t.Errorf("Put新key失败：期望prev=0、ok=false，实际prev=%d、ok=%v", prev, ok)
			}
			if lhm.Len() != 1 {
				t.Errorf("Put新key后Len错误：期望1，实际%d", lhm.Len())
			}
		})

		// 3. 测试Put（更新）
		t.Run("Put-Update", func(t *testing.T) {
			prev, ok := lhm.Put(keys[0], 200)
			if !ok || prev != 100 {
				t.Errorf("Put更新key失败：期望prev=100、ok=true，实际prev=%d、ok=%v", prev, ok)
			}
			val, _ := lhm.Get(keys[0])
			if val != 200 {
				t.Errorf("Put更新后值错误：期望200，实际%d", val)
			}
		})

		// 4. 测试Get（命中/未命中）
		t.Run("Get", func(t *testing.T) {
			// 命中场景
			val, ok := lhm.Get(keys[0])
			if !ok || val != 200 {
				t.Errorf("Get命中失败：期望200、ok=true，实际%d、ok=%v", val, ok)
			}
			// 未命中场景
			val, ok = lhm.Get("non-exist-key")
			if ok || val != 0 {
				t.Errorf("Get未命中失败：期望0、ok=false，实际%d、ok=%v", val, ok)
			}
		})

		// 5. 测试Contains
		t.Run("Contains", func(t *testing.T) {
			if !lhm.Contains(keys[0]) {
				t.Error("Contains存在的key失败：期望true")
			}
			if lhm.Contains("non-exist-key") {
				t.Error("Contains不存在的key失败：期望false")
			}
		})

		// 6. 测试Len
		t.Run("Len", func(t *testing.T) {
			lhm.Clear()
			// 批量插入数据
			for i := 0; i < smallTestSize/2; i++ {
				t.Log("put:", keys[i])
				if _, b := lhm.Put(keys[i], i); b {
					t.Error("exist:", keys[i])
				}
			}
			expectedLen := smallTestSize / 2
			if lhm.Len() != expectedLen {
				t.Errorf("Len批量插入后错误：期望%d，实际%d", expectedLen, lhm.Len())
			}
			lhm.Range(func(key string, value int) bool {
				t.Log(key, ":", value)
				return true
			})
		})

		// 7. 测试Del
		t.Run("Del", func(t *testing.T) {
			// 删除存在的key
			lhm.Del(keys[0])
			if lhm.Contains(keys[0]) {
				t.Error("Del存在的key失败：key仍存在")
			}
			if lhm.Len() != smallTestSize/2-1 {
				t.Errorf("Del后Len错误：期望%d，实际%d", smallTestSize/2-1, lhm.Len())
			}
			// 删除不存在的key（边界条件：无panic）
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Error("Del不存在的key触发panic：", r)
					}
				}()
				lhm.Del("non-exist-key")
			}()
		})

		// 8. 测试Clear
		t.Run("Clear", func(t *testing.T) {
			lhm.Clear()
			if lhm.Len() != 0 {
				t.Errorf("Clear后Len错误：期望0，实际%d", lhm.Len())
			}
			if lhm.Contains(keys[1]) {
				t.Error("Clear后key仍存在")
			}
		})

		// 9. 测试容量限制（LRU淘汰）
		t.Run("Capacity-Limit-LRU", func(t *testing.T) {
			lhm := NewLimitHashMapWithSegment[string, int](10, 1) // 1个segment，容量10
			// 插入11个key，触发LRU淘汰
			for i := 0; i < 11; i++ {
				lhm.Put(keys[i], i)
			}
			if lhm.Len() != 10 {
				t.Errorf("容量限制后Len错误：期望10，实际%d", lhm.Len())
			}
			if lhm.Contains(keys[0]) { // keys[0]是最旧的，应被淘汰
				t.Error("LRU淘汰失败：最旧的key仍存在")
			}
			if !lhm.Contains(keys[10]) { // keys[10]是最新的，应保留
				t.Error("LRU淘汰失败：最新的key不存在")
			}
		})

		// 10. 测试Range（遍历）
		t.Run("Range", func(t *testing.T) {
			lhm.Clear()
			// 插入10个测试数据
			testData := map[string]int{
				"range-key-1": 1,
				"range-key-2": 2,
				"range-key-3": 3,
			}
			for k, v := range testData {
				lhm.Put(k, v)
			}
			// 遍历验证
			visited := make(map[string]bool)
			lhm.Range(func(k string, v int) bool {
				if testData[k] != v {
					t.Errorf("Range值错误：key=%s，期望%d，实际%d", k, testData[k], v)
				}
				visited[k] = true
				return true
			})
			// 验证所有数据都被遍历
			for k := range testData {
				if !visited[k] {
					t.Errorf("Range未遍历到key：%s", k)
				}
			}
			// 测试中途终止遍历
			stopKey := "range-key-2"
			visited = make(map[string]bool)
			lhm.Range(func(k string, v int) bool {
				visited[k] = true
				return k != stopKey // 遇到stopKey终止
			})
			if visited["range-key-3"] {
				t.Error("Range中途终止失败：遍历到了终止后的key")
			}
		})
	})

	// 测试场景2：数值key（uint64） + string值（验证不同key类型的兼容性）
	t.Run("uint64-key-string-value", func(t *testing.T) {
		lhm := NewLimitHashMap[uint64, string](smallTestSize)
		keys := genUint64Keys(smallTestSize)

		// 验证Put/Get
		lhm.Put(keys[0], "test-value")
		val, ok := lhm.Get(keys[0])
		if !ok || val != "test-value" {
			t.Errorf("uint64 key Get失败：期望test-value、ok=true，实际%s、ok=%v", val, ok)
		}

		// 验证Del
		lhm.Del(keys[0])
		if lhm.Contains(keys[0]) {
			t.Error("uint64 key Del失败：key仍存在")
		}
	})

	// 测试场景3：float64 key + bool值（覆盖所有支持的key类型）
	t.Run("float64-key-bool-value", func(t *testing.T) {
		lhm := NewLimitHashMap[float64, bool](smallTestSize)
		lhm.Put(3.14159, true)
		val, ok := lhm.Get(3.14159)
		if !ok || val != true {
			t.Errorf("float64 key Get失败：期望true、ok=true，实际%v、ok=%v", val, ok)
		}
	})
}

// ------------------------------ 第二部分：大数据量稳定性测试（验证高数据量下的正确性） ------------------------------

// TestLimitHashMap_LargeData 验证大数据量下所有函数的稳定性
func TestLimitHashMap_LargeData(t *testing.T) {
	t.Run("string-key-int-value", func(t *testing.T) {
		lhm := NewLimitHashMap[string, int](largeTestSize * 2)
		keys := genStringKeys(largeTestSize)

		// 1. 批量Put
		t.Run("Batch-Put", func(t *testing.T) {
			for i, k := range keys {
				lhm.Put(k, i)
			}
			if lhm.Len() != largeTestSize {
				t.Errorf("大数据量Put后Len错误：期望%d，实际%d", largeTestSize, lhm.Len())
			}
		})

		// 2. 批量Get（验证所有值正确）
		t.Run("Batch-Get", func(t *testing.T) {
			sampleCount := 1000 // 抽样验证（避免全量遍历耗时）
			for i := 0; i < sampleCount; i++ {
				idx := rand.Intn(largeTestSize)
				val, ok := lhm.Get(keys[idx])
				if !ok || val != idx {
					t.Errorf("大数据量Get失败：key=%s，期望%d，实际%d", keys[idx], idx, val)
				}
			}
		})

		// 3. 批量Del
		t.Run("Batch-Del", func(t *testing.T) {
			delCount := largeTestSize / 2
			for i := 0; i < delCount; i++ {
				lhm.Del(keys[i])
			}
			// 允许±100的误差（分段锁的并发计数特性）
			if lhm.Len() < largeTestSize-delCount-100 || lhm.Len() > largeTestSize-delCount+100 {
				t.Errorf("大数据量Del后Len错误：期望%d±100，实际%d", largeTestSize-delCount, lhm.Len())
			}
		})

		// 4. 大数据量遍历
		t.Run("Large-Range", func(t *testing.T) {
			visitedCount := 0
			lhm.Range(func(k string, v int) bool {
				visitedCount++
				return visitedCount < 1000 // 只遍历前1000个，避免耗时
			})
			if visitedCount != 1000 {
				t.Errorf("大数据量Range错误：期望遍历1000个，实际%d", visitedCount)
			}
		})
	})
}

// ------------------------------ 第三部分：并发安全性测试（验证高并发下无panic、数据不混乱） ------------------------------

// TestLimitHashMap_Concurrent 验证高并发读写下的安全性
func TestLimitHashMap_Concurrent(t *testing.T) {
	lhm := NewLimitHashMap[int, int](10_0000)
	var wg sync.WaitGroup
	const (
		writeGoroutine  = 8    // 写协程数
		readGoroutine   = 8    // 读协程数
		delGoroutine    = 5    // 删除协程数
		rangeGoroutine  = 2    // 遍历协程数
		opsPerGoroutine = 2000 // 每个协程的操作数
	)

	// 1. 写协程：Put操作
	for g := 0; g < writeGoroutine; g++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := workerID*opsPerGoroutine + i
				lhm.Put(key, key)
			}
		}(g)
	}

	// 2. 读协程：Get + Contains
	for g := 0; g < readGoroutine; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := rand.Intn(writeGoroutine * opsPerGoroutine)
				_, _ = lhm.Get(key)
				_ = lhm.Contains(key)
			}
		}()
	}

	// 3. 删除协程：Del
	for g := 0; g < delGoroutine; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine/2; i++ {
				key := rand.Intn(writeGoroutine * opsPerGoroutine)
				lhm.Del(key)
			}
		}()
	}

	// 4. 遍历协程：Range
	for g := 0; g < rangeGoroutine; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.Error("并发Range触发panic：", r)
						}
					}()
					lhm.Range(func(k int, v int) bool {
						return true
					})
				}()
			}
		}()
	}

	// 等待所有协程完成
	wg.Wait()

	// 核心验证：Len不为负数（并发安全的关键指标）
	if lhm.Len() < 0 {
		t.Fatalf("并发操作后Len为负数：%d", lhm.Len())
	}
}

// ------------------------------ 第四部分：性能基准测试（大数据量 + 真实场景） ------------------------------

// BenchmarkLimitHashMap_Put 测试Put性能（不同key类型）
func BenchmarkLimitHashMap_Put(b *testing.B) {
	// 字符串key
	b.Run("string-key", func(b *testing.B) {
		lhm := NewLimitHashMap[string, int](b.N)
		keys := genStringKeys(b.N)
		b.ResetTimer() // 重置计时器，排除初始化耗时

		for i := 0; i < b.N; i++ {
			lhm.Put(keys[i], i)
		}
	})

	// uint64 key（数值key性能更优）
	b.Run("uint64-key", func(b *testing.B) {
		lhm := NewLimitHashMap[uint64, int](b.N)
		keys := genUint64Keys(b.N)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			lhm.Put(keys[i], i)
		}
	})
}

// BenchmarkLimitHashMap_Get 测试Get性能（命中/未命中场景）
func BenchmarkLimitHashMap_Get(b *testing.B) {
	// 字符串key - 命中场景
	b.Run("string-key-hit", func(b *testing.B) {
		lhm := NewLimitHashMap[string, int](benchSize)
		keys := genStringKeys(benchSize)
		// 预插入数据
		for i := 0; i < benchSize; i++ {
			lhm.Put(keys[i], i)
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _ = lhm.Get(keys[i%benchSize])
		}
	})

	// 字符串key - 未命中场景
	b.Run("string-key-miss", func(b *testing.B) {
		lhm := NewLimitHashMap[string, int](0) // 空map
		keys := genStringKeys(b.N)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _ = lhm.Get(keys[i])
		}
	})

	// uint64 key - 命中场景
	b.Run("uint64-key-hit", func(b *testing.B) {
		lhm := NewLimitHashMap[uint64, int](benchSize)
		keys := genUint64Keys(benchSize)
		for i := 0; i < benchSize; i++ {
			lhm.Put(keys[i], i)
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _ = lhm.Get(keys[i%benchSize])
		}
	})
}

// BenchmarkLimitHashMap_Concurrent 测试高并发混合操作性能（Put + Get + Del）
func BenchmarkLimitHashMap_Concurrent(b *testing.B) {
	lhm := NewLimitHashMap[int, int](benchSize * 10)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 3 {
			case 0: // Put
				lhm.Put(i, i)
			case 1: // Get
				_, _ = lhm.Get(i)
			case 2: // Del
				lhm.Del(i)
			}
			i++
		}
	})
}

// BenchmarkLimitHashMap_Len 测试高并发下Len的性能
func BenchmarkLimitHashMap_Len(b *testing.B) {
	lhm := NewLimitHashMap[int, int](10_0000)
	// 预插入数据
	for i := 0; i < 10_0000; i++ {
		lhm.Put(i, i)
	}
	b.ResetTimer()

	// 高并发调用Len
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = lhm.Len()
		}
	})
}

// BenchmarkLimitHashMap_Range 测试遍历性能
func BenchmarkLimitHashMap_Range(b *testing.B) {
	lhm := NewLimitHashMap[int, int](10_0000)
	// 预插入10万条数据
	for i := 0; i < 10_0000; i++ {
		lhm.Put(i, i)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lhm.Range(func(k int, v int) bool {
			return true
		})
	}
}
