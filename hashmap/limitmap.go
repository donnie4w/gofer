// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hashmap

import (
	"container/list"
	"sync"
)

// limitEntry represents a map entry with a generic key and value
type limitEntry[K any, V any] struct {
	key   K
	value V
}

// LimitMap represents a generic map with capacity limit
type LimitMap[K comparable, V any] struct {
	cache    map[any]*list.Element
	order    *list.List
	capacity int
	lock     sync.RWMutex
}

// NewLimitMap creates a new instance of LimitMap with the given capacity
func NewLimitMap[K comparable, V any](capacity int) *LimitMap[K, V] {
	return &LimitMap[K, V]{
		cache:    make(map[any]*list.Element),
		order:    list.New(),
		capacity: capacity,
	}
}

// Get retrieves a value from the cache by key
func (mc *LimitMap[K, V]) Get(key K) (value V, ok bool) {
	mc.lock.RLock()
	defer mc.lock.RUnlock()
	if elem, found := mc.cache[any(key)]; found {
		return elem.Value.(*limitEntry[K, V]).value, true
	}
	return
}

// Put adds or updates a key-value pair in the map
func (mc *LimitMap[K, V]) Put(key K, value V) (prev V, b bool) {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	if elem, exists := mc.cache[any(key)]; exists {
		prev, b = elem.Value.(*limitEntry[K, V]).value, true
		elem.Value.(*limitEntry[K, V]).value = value
		mc.order.MoveToFront(elem)
		return
	}
	if len(mc.cache) > mc.capacity {
		oldest := mc.order.Back()
		if oldest != nil {
			mc.order.Remove(oldest)
			delete(mc.cache, oldest.Value.(*limitEntry[K, V]).key)
		}
	}
	newEntry := &limitEntry[K, V]{key, value}
	elem := mc.order.PushFront(newEntry)
	mc.cache[any(key)] = elem
	return
}

// Del deletes a key-value pair from the map
func (mc *LimitMap[K, V]) Del(key K) {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	if elem, exists := mc.cache[any(key)]; exists {
		mc.order.Remove(elem)
		delete(mc.cache, any(key))
	}
}

// RemoveMulti deletes multiple key-value pairs from the map
func (mc *LimitMap[K, V]) RemoveMulti(keys []K) {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	for _, key := range keys {
		if elem, exists := mc.cache[any(key)]; exists {
			mc.order.Remove(elem)
			delete(mc.cache, any(key))
		}
	}
}

// Len length of map
func (mc *LimitMap[K, V]) Len() int {
	mc.lock.RLock()
	defer mc.lock.RUnlock()
	return len(mc.cache)
}
