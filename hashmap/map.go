// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/gofer
package hashmap

import (
	"sync"
	"sync/atomic"
)

type MapL[K any, V any] struct {
	m   sync.Map
	len int64
	mux *sync.Mutex
}

func NewMapL[K any, V any]() *MapL[K, V] {
	return &MapL[K, V]{m: sync.Map{}, mux: &sync.Mutex{}}
}

func (this *MapL[K, V]) Put(key K, value V) {
	if _, ok := this.m.Swap(key, value); !ok {
		atomic.AddInt64(&this.len, 1)
	}
}

func (this *MapL[K, V]) Get(key K) (_r V, b bool) {
	if v, ok := this.m.Load(key); ok {
		_r, b = v.(V), ok
	}
	return
}

func (this *MapL[K, V]) Has(key K) (ok bool) {
	_, ok = this.m.Load(key)
	return
}

func (this *MapL[K, V]) Del(key K) {
	this.mux.Lock()
	defer this.mux.Unlock()
	if _, ok := this.m.LoadAndDelete(key); ok {
		atomic.AddInt64(&this.len, -1)
	}
}

func (this *MapL[K, V]) Range(f func(k K, v V) bool) {
	this.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}

func (this *MapL[K, V]) Len() int64 {
	return this.len
}

/***********************************************************/
type Map[K any, V any] struct {
	m sync.Map
}

func NewMap[K any, V any]() *Map[K, V] {
	return &Map[K, V]{m: sync.Map{}}
}

func (this *Map[K, V]) Put(key K, value V) (v V, loaded bool) {
	if previous, ok := this.m.Swap(key, value); ok {
		v, loaded = previous.(V), ok
	}
	return
}

func (this *Map[K, V]) Get(key K) (_r V, b bool) {
	if v, ok := this.m.Load(key); ok {
		_r, b = v.(V), ok
	}
	return
}

func (this *Map[K, V]) Has(key K) (ok bool) {
	_, ok = this.m.Load(key)
	return
}

func (this *Map[K, V]) Del(key K) (v V, ok bool) {
	if e, loaded := this.m.LoadAndDelete(key); loaded {
		v, ok = e.(V), loaded
	}
	return
}

func (this *Map[K, V]) Range(f func(k K, v V) bool) {
	this.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}
