// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/gofer/cache

package cache

import (
	"sync"

	"github.com/golang/groupcache/lru"
)

type LruCache[T any] struct {
	cache *lru.Cache
	lock  *sync.Mutex
}

func NewLruCache[T any](maxEntries int) *LruCache[T] {
	return &LruCache[T]{cache: lru.New(maxEntries), lock: new(sync.Mutex)}
}

// Get retrieves a value from the cache. It uses RLock for concurrent reads.
func (lc *LruCache[T]) Get(key lru.Key) (value T, b bool) {
	lc.lock.Lock()
	defer lc.lock.Unlock()
	if v, ok := lc.cache.Get(key); ok {
		if v != nil {
			value = v.(T)
		}
		b = true
	}
	return
}

// Remove deletes an entry from the cache. It uses Lock for exclusive writes.
func (lc *LruCache[T]) Remove(key lru.Key) {
	lc.lock.Lock()
	defer lc.lock.Unlock()
	lc.cache.Remove(key)
}

// RemoveMulti deletes multiple entries from the cache. It uses Lock for exclusive writes.
func (lc *LruCache[T]) RemoveMulti(keys []string) {
	lc.lock.Lock()
	defer lc.lock.Unlock()

	for _, k := range keys {
		lc.cache.Remove(k)
	}
}

// Add adds a value to the cache. It uses Lock for exclusive writes.
func (lc *LruCache[T]) Add(key lru.Key, value T) {
	if key == nil {
		return
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()
	lc.cache.Add(key, value)
}
