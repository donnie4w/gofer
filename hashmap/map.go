// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/hashmap

package hashmap

import (
	"sync"
	"sync/atomic"
)

// MapL is a generic, thread-safe map with an atomic length counter.
// Built on sync.Map (lock-free for most operations).
// Supports nil values correctly (e.g. *T, []byte, map, etc.).
type MapL[K comparable, V any] struct {
	m   sync.Map
	len int64
}

// NewMapL creates a new MapL.
func NewMapL[K comparable, V any]() *MapL[K, V] {
	return &MapL[K, V]{}
}

// Put stores or updates the value for key.
// Returns the previous value and whether it existed.
// Correctly handles nil values (typed nil is preserved).
func (t *MapL[K, V]) Put(key K, value V) (prev V, existed bool) {
	if old, loaded := t.m.Swap(key, value); loaded {
		existed = true
		prev = old.(V)
	} else {
		atomic.AddInt64(&t.len, 1)
	}
	return
}

// Get returns the value for key and whether it exists.
// Correctly returns typed nil when the stored value is nil.
func (t *MapL[K, V]) Get(key K) (v V, ok bool) {
	if e, loaded := t.m.Load(key); loaded {
		v = e.(V)
		ok = true
	}
	return
}

// GetOrInsert retrieves the value associated with the specified key. If the key does not exist, it inserts the given value and returns that value.
// Return value: The value corresponding to the key, along with a flag indicating whether the value was newly inserted (false indicates it already existed, true indicates it was newly inserted).
func (t *MapL[K, V]) GetOrInsert(key K, value V) (v V, inserted bool) {
	if e, loaded := t.m.Load(key); loaded {
		v = e.(V)
		inserted = false
		return
	}

	if e, loaded := t.m.LoadOrStore(key, value); loaded {
		v = e.(V)
		inserted = false
	} else {
		v = value
		inserted = true
		atomic.AddInt64(&t.len, 1) // 维护长度计数器
	}
	return
}

// Has reports whether key exists.
func (t *MapL[K, V]) Has(key K) bool {
	_, ok := t.m.Load(key)
	return ok
}

// Del deletes the key and returns whether it existed.
func (t *MapL[K, V]) Del(key K) (existed bool) {
	if _, ok := t.m.LoadAndDelete(key); ok {
		atomic.AddInt64(&t.len, -1)
		existed = true
	}
	return
}

// Range calls f sequentially for each key/value pair.
// If f returns false, iteration stops.
// Correctly passes nil values to f.
func (t *MapL[K, V]) Range(f func(k K, v V) bool) {
	t.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}

// Len returns the number of elements (approximate under high contention).
func (t *MapL[K, V]) Len() int64 {
	return atomic.LoadInt64(&t.len)
}

// Clear removes all entries and resets length to 0.
func (t *MapL[K, V]) Clear() {
	t.m.Range(func(key, _ any) bool {
		t.m.Delete(key)
		return true
	})
	atomic.StoreInt64(&t.len, 0)
}

// Map is a generic, thread-safe map without length tracking.
// Maximum performance version.
type Map[K comparable, V any] struct {
	m sync.Map
}

// NewMap creates a new Map.
func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{}
}

// Put stores or updates the value for key.
// Returns the previous value and whether it existed.
func (t *Map[K, V]) Put(key K, value V) (prev V, existed bool) {
	if old, loaded := t.m.Swap(key, value); loaded {
		existed = true
		prev = old.(V)
	}
	return
}

// Get returns the value for key and whether it exists.
func (t *Map[K, V]) Get(key K) (v V, ok bool) {
	if e, loaded := t.m.Load(key); loaded {
		v = e.(V)
		ok = true
	}
	return
}

// GetOrInsert retrieves the value associated with the specified key. If the key does not exist, it inserts the given value and returns that value.
// Return value: The value corresponding to the key, along with a flag indicating whether the value was newly inserted (false indicates it already existed, true indicates it was newly inserted).
func (t *Map[K, V]) GetOrInsert(key K, value V) (v V, inserted bool) {
	if e, loaded := t.m.Load(key); loaded {
		v = e.(V)
		inserted = false
		return
	}

	if e, loaded := t.m.LoadOrStore(key, value); loaded {
		v = e.(V)
		inserted = false
	} else {
		v = value
		inserted = true
	}
	return
}

// Has reports whether key exists.
func (t *Map[K, V]) Has(key K) bool {
	_, ok := t.m.Load(key)
	return ok
}

// Del deletes the key and returns whether it existed.
func (t *Map[K, V]) Del(key K) bool {
	_, ok := t.m.LoadAndDelete(key)
	return ok
}

// Range calls f sequentially for each key/value pair.
func (t *Map[K, V]) Range(f func(k K, v V) bool) {
	t.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}

// Clear removes all entries.
func (t *Map[K, V]) Clear() {
	t.m.Range(func(key, _ any) bool {
		t.m.Delete(key)
		return true
	})
}
