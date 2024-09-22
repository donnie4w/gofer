// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/gofer/hashmap

package hashmap

import (
	"container/list"
	"hash/fnv"
	"math"
	"sync"
)

const numSegments = 1 << 6

type segment struct {
	cache map[uint64]*list.Element
	order *list.List
	mu    sync.RWMutex
}

// LimitHashMap is a capacity-limited hash map that supports different key types.
type LimitHashMap[K int | int64 | int8 | int16 | int32 | uint | uint64 | uint8 | uint16 | uint32 | float64 | float32 | uintptr | string, V any] struct {
	capacity      int
	segments      []*segment
	hashFunc      func(K) uint64
	segmentNumber int
}

type entry struct {
	hashedKey uint64
	key       any
	value     any
}

// NewLimitHashMap creates a new LimitHashMap with a specified capacity.
// Parameters:
//
//	capacity int - The total maximum capacity of the LimitHashMap.
//
// Returns:
//
//	*LimitHashMap[K, V] - A new instance of a capacity-limited LimitHashMap.
func NewLimitHashMap[K int | int64 | int8 | int16 | int32 | uint | uint64 | uint8 | uint16 | uint32 | float64 | float32 | uintptr | string, V any](capacity int) *LimitHashMap[K, V] {
	return NewLimitHashMapWithSegment[K, V](capacity, numSegments)
}

// NewLimitHashMapWithSegment creates a new LimitHashMap with a specified capacity and number of segments.
// Parameters:
//
//	capacity int - The total maximum capacity of the LimitHashMap.
//	segmentNumber int - The number of segments to divide the cache into.
//
// Returns:
//
//	*LimitHashMap[K, V] - A new instance of a capacity-limited LimitHashMap divided into segments.
func NewLimitHashMapWithSegment[K int | int64 | int8 | int16 | int32 | uint | uint64 | uint8 | uint16 | uint32 | float64 | float32 | uintptr | string, V any](capacity, segmentNumber int) *LimitHashMap[K, V] {
	segments := make([]*segment, segmentNumber)
	for i := 0; i < segmentNumber; i++ {
		segments[i] = &segment{
			cache: make(map[uint64]*list.Element),
			order: list.New(),
		}
	}

	var hashFunc func(K) uint64
	switch any(*new(K)).(type) {
	case uint64:
		hashFunc = func(key K) uint64 {
			return any(key).(uint64)
		}
	case uint32:
		hashFunc = func(key K) uint64 {
			return uint64(any(key).(uint32))
		}
	case uint16:
		hashFunc = func(key K) uint64 {
			return uint64(any(key).(uint16))
		}
	case int64:
		hashFunc = func(key K) uint64 {
			return uint64(any(key).(int64))
		}
	case int32:
		hashFunc = func(key K) uint64 {
			return uint64(any(key).(int32))
		}
	case int16:
		hashFunc = func(key K) uint64 {
			return uint64(any(key).(int16))
		}
	case int:
		hashFunc = func(key K) uint64 {
			return uint64(any(key).(int))
		}
	case uint:
		hashFunc = func(key K) uint64 {
			return uint64(any(key).(uint))
		}
	case float64:
		hashFunc = func(key K) uint64 {
			return math.Float64bits(any(key).(float64))
		}
	case float32:
		hashFunc = func(key K) uint64 {
			return math.Float64bits(float64(any(key).(float32)))
		}
	case int8:
		hashFunc = func(key K) uint64 {
			return uint64(any(key).(int8))
		}
	case uint8:
		hashFunc = func(key K) uint64 {
			return uint64(any(key).(uint8))
		}
	case uintptr:
		hashFunc = func(key K) uint64 {
			return uint64(any(key).(uintptr))
		}
	case string:
		hashFunc = func(key K) uint64 {
			h := fnv.New64a()
			h.Write([]byte(any(key).(string)))
			return h.Sum64()
		}
	default:
		panic("unsupported hashedKey type")
	}

	return &LimitHashMap[K, V]{
		capacity:      capacity / segmentNumber,
		segments:      segments,
		hashFunc:      hashFunc,
		segmentNumber: segmentNumber,
	}
}

func (c *LimitHashMap[K, V]) getSegment(key K) *segment {
	hashedKey := c.hashFunc(key)
	return c.segments[uint(hashedKey%uint64(c.segmentNumber))]
}

func (c *LimitHashMap[K, V]) Get(key K) (r V, b bool) {
	segment := c.getSegment(key)
	segment.mu.RLock()
	defer segment.mu.RUnlock()

	hashedKey := c.hashFunc(key)
	if ele, ok := segment.cache[hashedKey]; ok {
		if v := ele.Value.(*entry).value; v != nil {
			return v.(V), true
		}
	}
	return
}

func (c *LimitHashMap[K, V]) Put(key K, value V) (prev V, b bool) {
	segment := c.getSegment(key)
	segment.mu.Lock()
	defer segment.mu.Unlock()

	hashedKey := c.hashFunc(key)
	if ele, ok := segment.cache[hashedKey]; ok {
		prev, b = ele.Value.(*entry).value.(V), true
		ele.Value.(*entry).value = value
		return
	}

	if segment.order.Len() >= c.capacity {
		oldest := segment.order.Back()
		if oldest != nil {
			segment.order.Remove(oldest)
			delete(segment.cache, oldest.Value.(*entry).hashedKey)
		}
	}

	ele := segment.order.PushFront(&entry{hashedKey: hashedKey, key: key, value: value})
	segment.cache[hashedKey] = ele
	return
}

func (c *LimitHashMap[K, V]) Del(key K) {
	segment := c.getSegment(key)
	segment.mu.Lock()
	defer segment.mu.Unlock()

	hashedKey := c.hashFunc(key)
	if ele, ok := segment.cache[hashedKey]; ok {
		segment.order.Remove(ele)
		delete(segment.cache, hashedKey)
	}
}

func (c *LimitHashMap[K, V]) Contains(key K) bool {
	segment := c.getSegment(key)
	segment.mu.RLock()
	defer segment.mu.RUnlock()

	hashedKey := c.hashFunc(key)
	_, ok := segment.cache[hashedKey]
	return ok
}

func (c *LimitHashMap[K, V]) Clear() {
	for _, segment := range c.segments {
		segment.mu.Lock()
		segment.cache = make(map[uint64]*list.Element)
		segment.order.Init()
		segment.mu.Unlock()
	}
}

func (c *LimitHashMap[K, V]) Len() int {
	total := 0
	for _, segment := range c.segments {
		segment.mu.Lock()
		total += segment.order.Len()
		segment.mu.Unlock()
	}
	return total
}
