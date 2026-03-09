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

const defaultSegments = 1 << 6

var lruSeed = maphash.MakeSeed()

type entryLru[K comparable, V any] struct {
	prev, next *entryLru[K, V]
	key        K
	value      V
}

type segmentLru[K comparable, V any] struct {
	cache  map[K]*entryLru[K, V]
	head   *entryLru[K, V]
	tail   *entryLru[K, V]
	mu     sync.RWMutex
	length int32
	cap    int
}

func (s *segmentLru[K, V]) insertFront(e *entryLru[K, V]) {
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

func (s *segmentLru[K, V]) remove(e *entryLru[K, V]) {
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

func (s *segmentLru[K, V]) moveToFront(e *entryLru[K, V]) {
	if e == nil || s.head == e {
		return
	}
	s.remove(e)
	s.insertFront(e)
}

func (s *segmentLru[K, V]) evict() {
	if s.tail == nil {
		return
	}
	old := s.tail
	s.remove(old)
	delete(s.cache, old.key)
	atomic.AddInt32(&s.length, -1)
}

// LimitLruMap is a thread-safe, fixed-capacity LRU map.
// It uses sharded locking (multiple segments) for high concurrency.
// Key collisions are handled correctly by Go's native map[K].
// For string keys, maphash (concurrent-safe) is used for segment selection.
//
// By default Safe=false (fast path): Get does NOT promote on read (LRU updated only on Put).
// Call SetSafe() to enable strict LRU promotion on every Get (more correct but slightly slower).
type LimitLruMap[K comparable, V any] struct {
	segmentLrus []*segmentLru[K, V]
	mask        uint64
	hashFunc    func(K) uint64
	Safe        bool
}

// NewLimitLruMap creates a new LimitLruMap with the given total capacity.
// capacity should be >= 64 and a power of 2 for best performance (internal segments will be evenly distributed).
func NewLimitLruMap[K comparable, V any](capacity int) *LimitLruMap[K, V] {
	return NewLimitLruMapWithSegment[K, V](capacity, defaultSegments)
}

// NewLimitLruMapWithSegment creates a new LimitLruMap with custom segment count.
// capacity should be >= segmentLruNumber and a power of 2.
// segmentLruNumber must be a power of 2 (enforced).
func NewLimitLruMapWithSegment[K comparable, V any](capacity, segmentLruNumber int) *LimitLruMap[K, V] {
	if segmentLruNumber <= 0 || (segmentLruNumber&(segmentLruNumber-1)) != 0 {
		panic("segmentLruNumber must be positive power of 2")
	}

	segmentLrus := make([]*segmentLru[K, V], segmentLruNumber)
	// Each segment gets roughly equal capacity (total capacity is guaranteed)
	segCap := (capacity + segmentLruNumber - 1) / segmentLruNumber
	if segCap < 1 {
		segCap = 1
	}

	for i := 0; i < segmentLruNumber; i++ {
		segmentLrus[i] = &segmentLru[K, V]{
			cache:  make(map[K]*entryLru[K, V], segCap),
			cap:    segCap,
			length: 0,
		}
	}

	return &LimitLruMap[K, V]{
		segmentLrus: segmentLrus,
		mask:        uint64(segmentLruNumber - 1),
		hashFunc:    generateLruHashFunc[K](),
	}
}

// generateLruHashFunc returns a fast hash function used ONLY for selecting which segment to use.
// For string it uses maphash (concurrent-safe, low collision). Other types use direct cast for speed.
func generateLruHashFunc[K comparable]() func(K) uint64 {
	var k K
	switch any(k).(type) {
	case string:
		return func(key K) uint64 {
			return maphash.String(lruSeed, any(key).(string))
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
func (c *LimitLruMap[K, V]) getSegment(hashed uint64) *segmentLru[K, V] {
	return c.segmentLrus[hashed&c.mask]
}

// SetSafe enables strict LRU promotion on Get (move-to-front even on reads).
// Default (Safe=false) is faster and suitable for most scenarios (promotion only happens on Put).
// Returns the map itself for chaining.
func (c *LimitLruMap[K, V]) SetSafe() *LimitLruMap[K, V] {
	c.Safe = true
	return c
}

// Get returns the value for key and whether it exists.
// It is safe for concurrent use.
//
// Behavior depends on Safe flag:
//   - Safe=false (default): fast read-only path, NO move-to-front on Get.
//   - Safe=true: promotes entry to front on every successful Get (strict LRU).
func (c *LimitLruMap[K, V]) Get(key K) (V, bool) {
	var zero V
	hash := c.hashFunc(key)
	seg := c.getSegment(hash)

	if !c.Safe {
		// Fast path: no LRU promotion (most common use case)
		seg.mu.RLock()
		defer seg.mu.RUnlock()
		if e, ok := seg.cache[key]; ok {
			return e.value, true
		}
		return zero, false
	}

	// Safe path: double-check + upgrade to write lock to perform moveToFront
	seg.mu.RLock()
	_, ok := seg.cache[key]
	if !ok {
		seg.mu.RUnlock()
		return zero, false
	}
	seg.mu.RUnlock()

	seg.mu.Lock()
	defer seg.mu.Unlock()

	if e, ok := seg.cache[key]; ok {
		seg.moveToFront(e)
		return e.value, true
	}
	return zero, false
}

// Put inserts or updates the value for key and moves it to the front (most recently used).
// If the key already exists, the old value is returned.
// If capacity is full, the least recently used entry is evicted.
// Returns (oldValue, existed).
func (c *LimitLruMap[K, V]) Put(key K, value V) (V, bool) {
	var zero V
	hash := c.hashFunc(key)
	seg := c.getSegment(hash)

	seg.mu.Lock()
	defer seg.mu.Unlock()

	if e, ok := seg.cache[key]; ok {
		old := e.value
		e.value = value
		seg.moveToFront(e)
		return old, true
	}

	if int(atomic.LoadInt32(&seg.length)) >= seg.cap {
		seg.evict()
	}

	e := &entryLru[K, V]{
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
func (c *LimitLruMap[K, V]) Del(key K) {
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
func (c *LimitLruMap[K, V]) Len() int {
	total := 0
	for _, seg := range c.segmentLrus {
		total += int(atomic.LoadInt32(&seg.length))
	}
	return total
}

// Clear removes all entries.
// It locks each segment sequentially (brief inconsistency is acceptable for Clear).
func (c *LimitLruMap[K, V]) Clear() {
	for _, seg := range c.segmentLrus {
		seg.mu.Lock()
		seg.cache = make(map[K]*entryLru[K, V], seg.cap)
		seg.head = nil
		seg.tail = nil
		atomic.StoreInt32(&seg.length, 0)
		seg.mu.Unlock()
	}
}

// Contains reports whether key exists.
// It is safe for concurrent use (no LRU promotion).
func (c *LimitLruMap[K, V]) Contains(key K) bool {
	hash := c.hashFunc(key)
	seg := c.getSegment(hash)

	seg.mu.RLock()
	_, ok := seg.cache[key]
	seg.mu.RUnlock()

	return ok
}

// Range calls f for each key/value pair from most recently used to least recently used.
// If f returns false, iteration stops.
// It takes a snapshot per segment to minimize lock hold time.
func (c *LimitLruMap[K, V]) Range(f func(key K, value V) bool) {
	for _, seg := range c.segmentLrus {
		// Take a snapshot under read lock to avoid holding the lock during callback
		seg.mu.RLock()
		if atomic.LoadInt32(&seg.length) == 0 {
			seg.mu.RUnlock()
			continue
		}
		entries := make([]*entryLru[K, V], 0, seg.length)
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
