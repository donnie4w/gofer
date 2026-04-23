// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/hashmap

package hashmap

import (
	"container/list"
	"hash/maphash"
	"math"
	"sync"
	"sync/atomic"
)

const hashSegments = 1 << 8

var hashSeed = maphash.MakeSeed()

// hashSegment is a single shard of the map.
// It uses container/list for FIFO ordering and a map for O(1) lookups.
type hashSegment[K comparable, V any] struct {
	cache  map[K]*list.Element // key → list element (real key comparison, handles hash collisions)
	order  *list.List
	mu     sync.RWMutex
	length int32
}

// entry is the internal value stored in the linked list.
type entry[K any, V any] struct {
	key   K
	value V
}

// LimitHashMap is a capacity-limited FIFO map (PushFront + evict Back).
//
// Deprecated: use NewLimitFifoMap / NewLimitFifoMapWithSegment instead.
// This type is kept only for backward compatibility.
type LimitHashMap[K int | int64 | int8 | int16 | int32 | uint | uint64 | uint8 | uint16 | uint32 | float64 | float32 | uintptr | string, V any] struct {
	capacity      int
	segments      []*hashSegment[K, V]
	hashFunc      func(K) uint64
	segmentNumber int
}

// NewLimitHashMap creates a new LimitHashMap with a specified capacity.
//
// Deprecated: use NewLimitFifoMap instead.
func NewLimitHashMap[K int | int64 | int8 | int16 | int32 | uint | uint64 | uint8 | uint16 | uint32 | float64 | float32 | uintptr | string, V any](capacity int) *LimitHashMap[K, V] {
	return NewLimitHashMapWithSegment[K, V](capacity, hashSegments)
}

// NewLimitHashMapWithSegment creates a new LimitHashMap with a specified capacity and number of segments.
//
// Deprecated: use NewLimitFifoMapWithSegment instead.
func NewLimitHashMapWithSegment[K int | int64 | int8 | int16 | int32 | uint | uint64 | uint8 | uint16 | uint32 | float64 | float32 | uintptr | string, V any](capacity, segmentNumber int) *LimitHashMap[K, V] {
	segments := make([]*hashSegment[K, V], segmentNumber)

	// Each segment gets roughly equal capacity (total capacity is guaranteed)
	segCap := (capacity + segmentNumber - 1) / segmentNumber
	if segCap < 1 {
		segCap = 1
	}

	for i := 0; i < segmentNumber; i++ {
		segments[i] = &hashSegment[K, V]{
			cache:  make(map[K]*list.Element, segCap),
			order:  list.New(),
			length: 0,
		}
	}

	return &LimitHashMap[K, V]{
		capacity:      segCap,
		segments:      segments,
		hashFunc:      generateHashFunc[K](),
		segmentNumber: segmentNumber,
	}
}

// generateHashFunc returns a fast hash function used ONLY for selecting which segment to use.
// For string it uses maphash (concurrent-safe, low collision). Other types use direct cast for speed.
func generateHashFunc[K int | int64 | int8 | int16 | int32 | uint | uint64 | uint8 | uint16 | uint32 | float64 | float32 | uintptr | string]() func(K) uint64 {
	var k K
	switch any(k).(type) {
	case uint64:
		return func(key K) uint64 { return any(key).(uint64) }
	case uint32:
		return func(key K) uint64 { return uint64(any(key).(uint32)) }
	case uint16:
		return func(key K) uint64 { return uint64(any(key).(uint16)) }
	case uint8:
		return func(key K) uint64 { return uint64(any(key).(uint8)) }
	case int64:
		return func(key K) uint64 { return uint64(any(key).(int64)) }
	case int32:
		return func(key K) uint64 { return uint64(any(key).(int32)) }
	case int16:
		return func(key K) uint64 { return uint64(any(key).(int16)) }
	case int8:
		return func(key K) uint64 { return uint64(any(key).(int8)) }
	case int:
		return func(key K) uint64 { return uint64(any(key).(int)) }
	case uint:
		return func(key K) uint64 { return uint64(any(key).(uint)) }
	case float64:
		return func(key K) uint64 { return math.Float64bits(any(key).(float64)) }
	case float32:
		return func(key K) uint64 { return math.Float64bits(float64(any(key).(float32))) }
	case uintptr:
		return func(key K) uint64 { return uint64(any(key).(uintptr)) }
	case string:
		return func(key K) uint64 {
			return maphash.String(hashSeed, any(key).(string))
		}
	default:
		panic("unsupported key type")
	}
}

// getSegmentAndHash returns the shard and the hash (hash is used only for routing).
func (c *LimitHashMap[K, V]) getSegmentAndHash(key K) (*hashSegment[K, V], uint64) {
	hashedKey := c.hashFunc(key)
	segIdx := uint(hashedKey % uint64(c.segmentNumber))
	return c.segments[segIdx], hashedKey
}

// Get returns the value for key and whether it exists.
// It is safe for concurrent use.
func (c *LimitHashMap[K, V]) Get(key K) (r V, b bool) {
	segment, _ := c.getSegmentAndHash(key)
	segment.mu.RLock()
	defer segment.mu.RUnlock()

	if ele, ok := segment.cache[key]; ok {
		entry := ele.Value.(*entry[K, V])
		return entry.value, true
	}
	return
}

// Put inserts or updates the value for key.
// If the key already exists, the old value is returned.
// If capacity is full, the oldest entry is evicted.
// Returns (oldValue, existed).
func (c *LimitHashMap[K, V]) Put(key K, value V) (prev V, b bool) {
	segment, _ := c.getSegmentAndHash(key)
	segment.mu.Lock()
	defer segment.mu.Unlock()

	if ele, ok := segment.cache[key]; ok {
		entry := ele.Value.(*entry[K, V])
		prev, b = entry.value, true
		entry.value = value
		return
	}

	segmentLen := int(atomic.LoadInt32(&segment.length))
	if segmentLen >= c.capacity {
		oldest := segment.order.Back()
		if oldest != nil {
			oldEntry := oldest.Value.(*entry[K, V])
			segment.order.Remove(oldest)
			delete(segment.cache, oldEntry.key)
			atomic.AddInt32(&segment.length, -1)
		}
	}

	newEntry := &entry[K, V]{
		key:   key,
		value: value,
	}
	ele := segment.order.PushFront(newEntry)
	segment.cache[key] = ele
	atomic.AddInt32(&segment.length, 1)
	return
}

// Del removes the key if it exists.
// It is safe for concurrent use.
func (c *LimitHashMap[K, V]) Del(key K) {
	segment, _ := c.getSegmentAndHash(key)
	segment.mu.Lock()
	defer segment.mu.Unlock()

	if ele, ok := segment.cache[key]; ok {
		segment.order.Remove(ele)
		delete(segment.cache, key)
		atomic.AddInt32(&segment.length, -1)
	}
}

// Contains reports whether key exists.
// It is safe for concurrent use.
func (c *LimitHashMap[K, V]) Contains(key K) bool {
	segment, _ := c.getSegmentAndHash(key)
	segment.mu.RLock()
	defer segment.mu.RUnlock()

	_, ok := segment.cache[key]
	return ok
}

// Clear removes all entries.
// It locks each segment sequentially.
func (c *LimitHashMap[K, V]) Clear() {
	for _, segment := range c.segments {
		segment.mu.Lock()
		segment.cache = make(map[K]*list.Element, c.capacity)
		segment.order.Init()
		atomic.StoreInt32(&segment.length, 0)
		segment.mu.Unlock()
	}
}

// Len returns the total number of entries (sum across all segments).
// It is safe for concurrent use and uses atomic reads.
func (c *LimitHashMap[K, V]) Len() int {
	total := 0
	for _, segment := range c.segments {
		total += int(atomic.LoadInt32(&segment.length))
	}
	return total
}

// Range iterates over all key-value pairs in FIFO order (oldest to newest).
// The iteration stops immediately when the provided function returns false.
// It takes a snapshot per segment to minimize lock hold time (writes are not blocked during callback).
func (c *LimitHashMap[K, V]) Range(f func(key K, value V) bool) {
	for _, seg := range c.segments {
		seg.mu.RLock()
		if atomic.LoadInt32(&seg.length) == 0 {
			seg.mu.RUnlock()
			continue
		}
		snapshot := make([]*entry[K, V], 0, int(atomic.LoadInt32(&seg.length)))
		for e := seg.order.Front(); e != nil; e = e.Next() {
			snapshot = append(snapshot, e.Value.(*entry[K, V]))
		}
		seg.mu.RUnlock()

		for _, entry := range snapshot {
			if !f(entry.key, entry.value) {
				return
			}
		}
	}
}
