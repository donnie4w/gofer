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

func (this *LruCache[T]) Get(key lru.Key) (value T, b bool) {
	defer recover()
	defer this.lock.Unlock()
	this.lock.Lock()
	if v, ok := this.cache.Get(key); ok {
		if v != nil {
			value = v.(T)
		}
		b = true
	}
	return
}

func (this *LruCache[T]) Remove(key lru.Key) {
	defer recover()
	defer this.lock.Unlock()
	this.lock.Lock()
	this.cache.Remove(key)
}

func (this *LruCache[T]) RemoveMulti(keys []string) {
	defer recover()
	defer this.lock.Unlock()
	this.lock.Lock()
	for _, k := range keys {
		this.cache.Remove(k)
	}
}

func (this *LruCache[T]) Add(key lru.Key, value T) {
	defer recover()
	defer this.lock.Unlock()
	this.lock.Lock()
	this.cache.Add(key, value)
}
