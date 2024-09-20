package hashmap

import (
	"container/list"
	"sync"
)

// limitEntry represents a cache entry with a generic key and value
type limitEntry[K any, V any] struct {
	key   K
	value V
}

// LimitMap represents a generic cache with capacity limit
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

// Put adds or updates a key-value pair in the cache
func (mc *LimitMap[K, V]) Put(key K, value V) {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	if elem, exists := mc.cache[any(key)]; exists {
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
}

// Remove deletes a key-value pair from the cache
func (mc *LimitMap[K, V]) Remove(key K) {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	if elem, exists := mc.cache[any(key)]; exists {
		mc.order.Remove(elem)
		delete(mc.cache, any(key))
	}
}

// RemoveMulti deletes multiple key-value pairs from the cache
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
