// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/hashmap

package hashmap

import (
	"fmt"
	"hash/maphash"
	"math"
	"sync"
	"sync/atomic"
)

const fifoSegments = 1 << 6

// fifoSeed is a global, read-only seed used for maphash.
// It is initialized once at package load time and is safe for concurrent use across all goroutines.
var fifoSeed = maphash.MakeSeed()

// entryFifo is the internal doubly-linked list node for the FIFO order.
type entryFifo[K comparable, V any] struct {
	prev, next *entryFifo[K, V]
	key        K
	value      V
}

// segmentFifo is a single shard of the map.
// Each segment has its own RWMutex, cache, and FIFO linked list to reduce lock contention.
type segmentFifo[K comparable, V any] struct {
	cache  map[K]*entryFifo[K, V] // key → entry (real key comparison, handles hash collisions)
	head   *entryFifo[K, V]
	tail   *entryFifo[K, V]
	mu     sync.RWMutex
	length int32
	cap    int
}

// insertFront adds the entry to the front of the FIFO list (most recently inserted).
func (s *segmentFifo[K, V]) insertFront(e *entryFifo[K, V]) {
	e.prev = nil
	e.next = s.head
	if s.head != nil {
		s.head.prev = e
	}
	s.head = e
	if s.tail == nil {
		s.tail = e
	}
}

func (s *segmentFifo[K, V]) remove(e *entryFifo[K, V]) {
	if e == nil {
		return
	}
	if e.prev != nil {
		e.prev.next = e.next
	} else {
		s.head = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	} else {
		s.tail = e.prev
	}
	e.prev = nil
	e.next = nil
}

// evict removes the oldest entry (tail) when capacity is reached.
func (s *segmentFifo[K, V]) evict() {
	if s.tail == nil {
		return
	}
	old := s.tail
	s.remove(old)
	delete(s.cache, old.key)
	atomic.AddInt32(&s.length, -1)
}

// LimitFifoMap is a thread-safe, fixed-capacity FIFO map.
// It uses sharded locking (multiple segments) for high concurrency.
type LimitFifoMap[K comparable, V any] struct {
	segmentFifos []*segmentFifo[K, V]
	mask         uint64
	hashFunc     func(K) uint64
}

// NewLimitFifoMap creates a new LimitFifoMap with the given total capacity.
// capacity should be >= 64 and a power of 2 for best performance (internal segments will be evenly distributed).
func NewLimitFifoMap[K comparable, V any](capacity int) *LimitFifoMap[K, V] {
	return NewLimitFifoMapWithSegment[K, V](capacity, fifoSegments)
}

// NewLimitFifoMapWithSegment creates a new LimitFifoMap with custom segment count.
// capacity should be >= segmentFifoNumber and a power of 2.
// segmentFifoNumber must be a power of 2 (enforced).
func NewLimitFifoMapWithSegment[K comparable, V any](capacity, segmentFifoNumber int) *LimitFifoMap[K, V] {
	if segmentFifoNumber <= 0 || (segmentFifoNumber&(segmentFifoNumber-1)) != 0 {
		panic("segmentFifoNumber must be positive power of 2")
	}

	segmentFifos := make([]*segmentFifo[K, V], segmentFifoNumber)
	// Each segment gets roughly equal capacity (total capacity is guaranteed)
	segCap := (capacity + segmentFifoNumber - 1) / segmentFifoNumber
	if segCap < 1 {
		segCap = 1
	}

	for i := 0; i < segmentFifoNumber; i++ {
		segmentFifos[i] = &segmentFifo[K, V]{
			cache:  make(map[K]*entryFifo[K, V], segCap),
			cap:    segCap,
			length: 0,
		}
	}

	return &LimitFifoMap[K, V]{
		segmentFifos: segmentFifos,
		mask:         uint64(segmentFifoNumber - 1),
		hashFunc:     generateFifoHashFunc[K](),
	}
}

// generateFifoHashFunc returns a fast hash function used ONLY for selecting which segment to use.
func generateFifoHashFunc[K comparable]() func(K) uint64 {
	var k K
	switch any(k).(type) {
	case string:
		return func(key K) uint64 {
			return maphash.String(fifoSeed, any(key).(string))
		}
	case int:
		return func(key K) uint64 { return uint64(any(key).(int)) }
	case int8:
		return func(key K) uint64 { return uint64(any(key).(int8)) }
	case int16:
		return func(key K) uint64 { return uint64(any(key).(int16)) }
	case int32:
		return func(key K) uint64 { return uint64(any(key).(int32)) }
	case int64:
		return func(key K) uint64 { return uint64(any(key).(int64)) }
	case uint:
		return func(key K) uint64 { return uint64(any(key).(uint)) }
	case uint8:
		return func(key K) uint64 { return uint64(any(key).(uint8)) }
	case uint16:
		return func(key K) uint64 { return uint64(any(key).(uint16)) }
	case uint32:
		return func(key K) uint64 { return uint64(any(key).(uint32)) }
	case uint64:
		return func(key K) uint64 { return any(key).(uint64) }
	case uintptr:
		return func(key K) uint64 { return uint64(any(key).(uintptr)) }
	case float32:
		return func(key K) uint64 { return math.Float64bits(float64(any(key).(float32))) }
	case float64:
		return func(key K) uint64 { return math.Float64bits(any(key).(float64)) }
	default:
		panic(fmt.Sprintf("unsupported key type: %T", k))
	}
}

// getSegment returns the shard responsible for the given hash (used only for routing).
func (c *LimitFifoMap[K, V]) getSegment(hashed uint64) *segmentFifo[K, V] {
	return c.segmentFifos[hashed&c.mask]
}

// Get returns the value for key and whether it exists.
// It is safe for concurrent use.
func (c *LimitFifoMap[K, V]) Get(key K) (V, bool) {
	var zero V
	hash := c.hashFunc(key)
	seg := c.getSegment(hash)

	seg.mu.RLock()
	defer seg.mu.RUnlock()

	if e, ok := seg.cache[key]; ok {
		return e.value, true
	}
	return zero, false
}

// Put inserts or updates the value for key.
// If the key already exists, the old value is returned and the entry is NOT moved (pure FIFO).
// If capacity is full, the oldest entry is evicted.
// Returns (oldValue, existed).
func (c *LimitFifoMap[K, V]) Put(key K, value V) (V, bool) {
	var zero V
	hash := c.hashFunc(key)
	seg := c.getSegment(hash)

	seg.mu.Lock()
	defer seg.mu.Unlock()

	if e, ok := seg.cache[key]; ok {
		old := e.value
		e.value = value
		return old, true
	}

	if int(atomic.LoadInt32(&seg.length)) >= seg.cap {
		seg.evict()
	}

	e := &entryFifo[K, V]{
		key:   key,
		value: value,
	}
	seg.insertFront(e)
	seg.cache[key] = e
	atomic.AddInt32(&seg.length, 1)

	return zero, false
}

// Del removes the key if it exists.
// It is safe for concurrent use.
func (c *LimitFifoMap[K, V]) Del(key K) {
	hash := c.hashFunc(key)
	seg := c.getSegment(hash)

	seg.mu.Lock()
	defer seg.mu.Unlock()

	if e, ok := seg.cache[key]; ok {
		seg.remove(e)
		delete(seg.cache, key)
		atomic.AddInt32(&seg.length, -1)
	}
}

// Len returns the total number of entries (sum across all segments).
// It is safe for concurrent use and uses atomic reads.
func (c *LimitFifoMap[K, V]) Len() int {
	total := 0
	for _, seg := range c.segmentFifos {
		total += int(atomic.LoadInt32(&seg.length))
	}
	return total
}

// Clear removes all entries.
// It locks each segment sequentially (brief inconsistency is acceptable for Clear).
func (c *LimitFifoMap[K, V]) Clear() {
	for _, seg := range c.segmentFifos {
		seg.mu.Lock()
		seg.cache = make(map[K]*entryFifo[K, V], seg.cap)
		seg.head = nil
		seg.tail = nil
		atomic.StoreInt32(&seg.length, 0)
		seg.mu.Unlock()
	}
}

// Contains reports whether key exists.
// It is safe for concurrent use.
func (c *LimitFifoMap[K, V]) Contains(key K) bool {
	hash := c.hashFunc(key)
	seg := c.getSegment(hash)

	seg.mu.RLock()
	_, ok := seg.cache[key]
	seg.mu.RUnlock()

	return ok
}

// Range calls f for each key/value pair in insertion order (oldest to newest).
// If f returns false, iteration stops.
// It takes a snapshot per segment to minimize lock hold time.
func (c *LimitFifoMap[K, V]) Range(f func(key K, value V) bool) {
	for _, seg := range c.segmentFifos {
		// Take a snapshot under read lock to avoid holding the lock during callback
		seg.mu.RLock()
		if atomic.LoadInt32(&seg.length) == 0 {
			seg.mu.RUnlock()
			continue
		}
		entries := make([]*entryFifo[K, V], 0, seg.length)
		for e := seg.head; e != nil; e = e.next {
			entries = append(entries, e)
		}
		seg.mu.RUnlock()

		for _, e := range entries {
			if !f(e.key, e.value) {
				return
			}
		}
	}
}
