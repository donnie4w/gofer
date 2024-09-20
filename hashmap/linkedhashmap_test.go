package hashmap

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkParallelLinkedHashMap(b *testing.B) {
	linkmap := NewLinkedHashMap[int64, int64](1 << 32)
	b.ResetTimer()
	i := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i++
			k := time.Now().UnixNano() + int64(i)
			linkmap.Put(k, time.Now().UnixNano())
			if k%9 == 0 {
				linkmap.Delete(k)
			}
			linkmap.Get(k)
		}
	})
}

func BenchmarkParallelLinkedHashMap2(b *testing.B) {
	linkmap := NewLinkedHashMap[int64, int64](1 << 32)
	b.ResetTimer()
	i := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i++
			k := time.Now().UnixNano() + int64(i)
			linkmap.Put(k, time.Now().UnixNano())
			if k%3 == 0 {
				iter := linkmap.Iterator(false)
				for {
					if key, _, ok := iter.Next(); ok && key%5 == 0 {
						linkmap.Delete(key)
					} else {
						break
					}
				}
			}
		}
	})
	fmt.Println(linkmap.Len())
}

func BenchmarkSerialLinkedHashMap(b *testing.B) {
	linkmap := NewLinkedHashMap[int64, int64](1 << 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := time.Now().UnixNano() + int64(i)
		linkmap.Put(k, time.Now().UnixNano())
		if k%3 == 0 {
			linkmap.Delete(k)
		}
		linkmap.Get(k)
	}
}

func TestLinkedHashMap(t *testing.T) {
	linkmap := NewLinkedHashMap[string, string](1 << 15)
	linkmap.Put("key1", "value1")
	linkmap.Put("key2", "value2")
	linkmap.Put("key3", "value3")
	linkmap.Put("key4", "value4")
	linkmap.Put("key5", "value5")
	linkmap.Put("key6", "value6")
	linkmap.Put("key7", "value7")
	fmt.Println(linkmap.Front())
	fmt.Println(linkmap.Back())
	linkmap.Delete("key2")
	linkmap.MoveToFront("key1")
	iter := linkmap.Iterator(false)
	for {
		key, value, ok := iter.Next()
		if !ok {
			break
		}
		fmt.Printf("Key: %s, Value: %s\n", key, value)
	}
	println()
	iter = linkmap.Iterator(true)
	for {
		key, value, ok := iter.Next()
		if !ok {
			break
		}
		fmt.Printf("Key: %s, Value: %s\n", key, value)
	}
}
