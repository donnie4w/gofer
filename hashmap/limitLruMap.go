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

const defaultSegments = 1 << 8

var lruSeed = maphash.MakeSeed()

type entryLru[K comparable, V any] struct {
	prev, next *entryLru[K, V]
	key        K
	value      V
	hits       int32 // atomic counter for sampled promotion
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

// LRUMode defines the strategy for updating access order in the cache.
type LRUMode int

const (
	// ModePerfFirst Performance first: sampled update order (default).
	// Moves an entry to the head only every GetsPerPromote Get operations.
	ModePerfFirst LRUMode = iota
	// ModeStrictLRU Strict LRU: moves an entry to the head on every successful Get.
	// More accurate but incurs more write locks.
	ModeStrictLRU
)

// LimitLruMap is a thread-safe, fixed-capacity LRU map.
// It uses segmented locking (multiple segments) to support high concurrency.
// By default, it uses performance-first mode (sampled promotion).
// Switch to strict LRU by using WithMode(ModeStrictLRU).
type LimitLruMap[K comparable, V any] struct {
	segmentLrus    []*segmentLru[K, V]
	mask           uint64
	hashFunc       func(K) uint64
	mode           LRUMode
	getsPerPromote int32 // Only effective in ModePerfFirst, default is 16
}

// NewLimitLruMap creates a new LimitLruMap with total capacity.
// It is recommended that capacity >= 64 and is a power of 2.
func NewLimitLruMap[K comparable, V any](capacity int) *LimitLruMap[K, V] {
	return NewLimitLruMapWithSegment[K, V](capacity, defaultSegments)
}

// NewLimitLruMapWithSegment creates a LimitLruMap with a custom number of segments.
func NewLimitLruMapWithSegment[K comparable, V any](capacity, segmentLruNumber int) *LimitLruMap[K, V] {
	if segmentLruNumber <= 0 || (segmentLruNumber&(segmentLruNumber-1)) != 0 {
		panic("segmentLruNumber must be positive power of 2")
	}

	segmentLrus := make([]*segmentLru[K, V], segmentLruNumber)
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
		segmentLrus:    segmentLrus,
		mask:           uint64(segmentLruNumber - 1),
		hashFunc:       generateLruHashFunc[K](),
		mode:           ModePerfFirst, // Default: performance first
		getsPerPromote: 16,            // Default: try promotion every 16 Gets
	}
}

// WithMode sets the LRU update mode (for method chaining).
func (c *LimitLruMap[K, V]) WithMode(mode LRUMode) *LimitLruMap[K, V] {
	c.mode = mode
	return c
}

// WithGetsPerPromote sets the sampling frequency (only effective in ModePerfFirst, recommended range: 4~64).
func (c *LimitLruMap[K, V]) WithGetsPerPromote(n int) *LimitLruMap[K, V] {
	if n < 1 {
		n = 1
	}
	c.getsPerPromote = int32(n)
	return c
}

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

func (c *LimitLruMap[K, V]) getSegment(hashed uint64) *segmentLru[K, V] {
	return c.segmentLrus[hashed&c.mask]
}

// Get returns the value for a given key and whether it exists.
// Updates the LRU order based on the configured Mode.
func (c *LimitLruMap[K, V]) Get(key K) (V, bool) {
	var zero V
	hash := c.hashFunc(key)
	seg := c.getSegment(hash)

	seg.mu.RLock()
	e, ok := seg.cache[key]
	if !ok {
		seg.mu.RUnlock()
		return zero, false
	}

	if c.mode == ModePerfFirst {
		hits := atomic.AddInt32(&e.hits, 1)
		shouldPromote := c.getsPerPromote > 0 && (hits%c.getsPerPromote == 0)
		seg.mu.RUnlock()

		if shouldPromote {
			seg.mu.Lock()
			if e2, stillOk := seg.cache[key]; stillOk && e2 == e {
				seg.moveToFront(e)
				atomic.StoreInt32(&e.hits, 0)
			}
			seg.mu.Unlock()
		}
	} else {
		// ModeStrictLRU: attempt to move to front on every access
		seg.mu.RUnlock()

		seg.mu.Lock()
		if e2, stillOk := seg.cache[key]; stillOk && e2 == e {
			seg.moveToFront(e)
		}
		seg.mu.Unlock()
	}

	return e.value, true
}

// Put inserts or updates a key-value pair and moves it to the head.
// If capacity is exceeded, evicts the least recently used element.
func (c *LimitLruMap[K, V]) Put(key K, value V) (old V, existed bool) {
	var zero V
	hash := c.hashFunc(key)
	seg := c.getSegment(hash)

	seg.mu.Lock()
	defer seg.mu.Unlock()

	if e, ok := seg.cache[key]; ok {
		old = e.value
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
		// hits defaults to 0
	}
	seg.insertFront(e)
	seg.cache[key] = e
	atomic.AddInt32(&seg.length, 1)

	return zero, false
}

// Del deletes a key if it exists.
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

// Len returns the approximate total number of elements currently in the map.
func (c *LimitLruMap[K, V]) Len() int {
	total := 0
	for _, seg := range c.segmentLrus {
		total += int(atomic.LoadInt32(&seg.length))
	}
	return total
}

// Clear removes all entries from the map.
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

// Contains checks if a key exists without updating the LRU order.
func (c *LimitLruMap[K, V]) Contains(key K) bool {
	hash := c.hashFunc(key)
	seg := c.getSegment(hash)

	seg.mu.RLock()
	_, ok := seg.cache[key]
	seg.mu.RUnlock()
	return ok
}

// Range iterates over all entries from most recently used to least recently used.
// Note: Each segment is snapshotted independently during iteration.
func (c *LimitLruMap[K, V]) Range(f func(key K, value V) bool) {
	for _, seg := range c.segmentLrus {
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
