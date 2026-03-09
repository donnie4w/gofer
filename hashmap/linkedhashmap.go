// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/hashmap

package hashmap

import (
	"container/list"
	"sync"
	"sync/atomic"
)

// LinkedHashMap is a thread-safe map that preserves insertion order (or manual MoveToFront order).
// It has a capacity limit: when full, the oldest entry is automatically evicted on Put/LoadOrStore.
// Read (Get) is lock-free; write operations use a single mutex (order maintenance requires it).
// Supports nil values correctly (e.g. *T, []byte, map, etc.).
type LinkedHashMap[K comparable, V any] struct {
	m     sync.Map
	list  *list.List
	limit int64
	count int64
	mu    sync.Mutex
}

// NewLinkedHashMap creates a new LinkedHashMap with capacity limit.
// When size exceeds limit, the oldest entry is evicted.
func NewLinkedHashMap[K comparable, V any](limit int64) *LinkedHashMap[K, V] {
	if limit <= 0 {
		limit = 1 << 10 // default 1024 if invalid
	}
	return &LinkedHashMap[K, V]{
		list:  list.New(),
		limit: limit,
	}
}

type entrykv[K, V any] struct {
	key   K
	value V
}

// putInternal is the shared logic for Put and LoadOrStore.
func (t *LinkedHashMap[K, V]) putInternal(k K, v V) (*entrykv[K, V], *list.Element, bool) {
	kv := &entrykv[K, V]{key: k, value: v}
	elem := t.list.PushFront(kv)

	if pre, ok := t.m.Swap(k, elem); ok && pre != nil {
		// existed → remove old list element
		t.list.Remove(pre.(*list.Element))
		oldEntry := pre.(*list.Element).Value.(*entrykv[K, V])
		return oldEntry, elem, true
	}

	// new entry
	atomic.AddInt64(&t.count, 1)
	if atomic.LoadInt64(&t.count) > t.limit {
		if oldest := t.list.Back(); oldest != nil {
			t.list.Remove(oldest)
			t.m.LoadAndDelete(oldest.Value.(*entrykv[K, V]).key)
			atomic.AddInt64(&t.count, -1)
		}
	}
	return nil, elem, false
}

// Put stores or updates the value and moves it to the front.
// If capacity is exceeded, the oldest entry is evicted.
func (t *LinkedHashMap[K, V]) Put(k K, v V) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.putInternal(k, v)
}

// LoadOrStore is like sync.Map.LoadOrStore but also maintains order and capacity.
// Returns the actual value stored and whether it was loaded (existed).
// If existed: returns old value + true, and updates to new v + moves to front.
// If not existed: returns the provided v + false.
func (t *LinkedHashMap[K, V]) LoadOrStore(k K, v V) (actual V, loaded bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	oldEntry, _, existed := t.putInternal(k, v)
	if existed {
		actual = oldEntry.value
		loaded = true
	} else {
		actual = v
		loaded = false
	}
	return
}

// Get returns the value and whether it exists (lock-free read).
func (t *LinkedHashMap[K, V]) Get(k K) (v V, ok bool) {
	if e, loaded := t.m.Load(k); loaded {
		entry := e.(*list.Element).Value.(*entrykv[K, V])
		v = entry.value
		ok = true
	}
	return
}

// Delete removes the key if it exists.
func (t *LinkedHashMap[K, V]) Delete(k K) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if elem, ok := t.m.LoadAndDelete(k); ok {
		t.list.Remove(elem.(*list.Element))
		atomic.AddInt64(&t.count, -1)
	}
}

// MoveToFront moves the entry to the most recent position (useful for manual LRU).
func (t *LinkedHashMap[K, V]) MoveToFront(k K) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if elem, ok := t.m.Load(k); ok {
		t.list.MoveToFront(elem.(*list.Element))
	}
}

// Back returns the oldest entry (least recently inserted/promoted).
func (t *LinkedHashMap[K, V]) Back() (k K, v V, ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if e := t.list.Back(); e != nil {
		entry := e.Value.(*entrykv[K, V])
		return entry.key, entry.value, true
	}
	return
}

// Front returns the newest entry (most recently inserted/promoted).
func (t *LinkedHashMap[K, V]) Front() (k K, v V, ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if e := t.list.Front(); e != nil {
		entry := e.Value.(*entrykv[K, V])
		return entry.key, entry.value, true
	}
	return
}

// Len returns the current number of entries (lock-free, atomic).
func (t *LinkedHashMap[K, V]) Len() int64 {
	return atomic.LoadInt64(&t.count)
}

// Clear removes all entries.
func (t *LinkedHashMap[K, V]) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.m = sync.Map{}
	t.list.Init()
	atomic.StoreInt64(&t.count, 0)
}

// Range iterates from most recent to oldest.
// Stops if f returns false.
// Snapshot-based: safe even if map is modified during iteration.
func (t *LinkedHashMap[K, V]) Range(f func(k K, v V) bool) {
	t.mu.Lock()
	snapshot := make([]*entrykv[K, V], 0, int(atomic.LoadInt64(&t.count)))
	for e := t.list.Front(); e != nil; e = e.Next() {
		snapshot = append(snapshot, e.Value.(*entrykv[K, V]))
	}
	t.mu.Unlock()

	for _, e := range snapshot {
		if !f(e.key, e.value) {
			return
		}
	}
}

// LinkedHashMapIterator provides ordered iteration (snapshot-based, no lock during Next).
type LinkedHashMapIterator[K, V any] struct {
	entries []*entrykv[K, V]
	index   int
}

// Next returns the next key/value pair.
// Returns false when iteration ends.
func (it *LinkedHashMapIterator[K, V]) Next() (K, V, bool) {
	if it.index < len(it.entries) {
		e := it.entries[it.index]
		it.index++
		return e.key, e.value, true
	}
	var k K
	var v V
	return k, v, false
}

// Iterator returns an iterator starting from front (true = newest first) or back (false = oldest first).
// The snapshot is taken at creation time, so it is safe and consistent even if the map changes.
func (t *LinkedHashMap[K, V]) Iterator(front bool) *LinkedHashMapIterator[K, V] {
	t.mu.Lock()
	defer t.mu.Unlock()

	snapshot := make([]*entrykv[K, V], 0, int(atomic.LoadInt64(&t.count)))

	if front {
		for e := t.list.Front(); e != nil; e = e.Next() {
			snapshot = append(snapshot, e.Value.(*entrykv[K, V]))
		}
	} else {
		for e := t.list.Back(); e != nil; e = e.Prev() {
			snapshot = append(snapshot, e.Value.(*entrykv[K, V]))
		}
	}

	return &LinkedHashMapIterator[K, V]{entries: snapshot}
}
