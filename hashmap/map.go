// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/gofer/hashmap

package hashmap

import (
	"sync"
	"sync/atomic"
)

type MapL[K any, V any] struct {
	m   sync.Map
	len int64
}

func NewMapL[K any, V any]() *MapL[K, V] {
	return &MapL[K, V]{m: sync.Map{}}
}

func (this *MapL[K, V]) Put(key K, value V) {
	if _, ok := this.m.Swap(key, value); !ok {
		atomic.AddInt64(&this.len, 1)
	}
}

func (this *MapL[K, V]) Get(key K) (_r V, b bool) {
	if v, ok := this.m.Load(key); ok {
		if v != nil {
			_r = v.(V)
		}
		b = true
	}
	return
}

func (this *MapL[K, V]) Has(key K) (ok bool) {
	_, ok = this.m.Load(key)
	return
}

func (this *MapL[K, V]) Del(key K) (ok bool) {
	if _, ok = this.m.LoadAndDelete(key); ok {
		atomic.AddInt64(&this.len, -1)
	}
	return
}

func (this *MapL[K, V]) Range(f func(k K, v V) bool) {
	this.m.Range(func(k, v any) bool {
		if v != nil {
			return f(k.(K), v.(V))
		} else {
			var t V
			return f(k.(K), t)
		}
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

func (this *Map[K, V]) Put(key K, value V) {
	this.m.Swap(key, value)
}

func (this *Map[K, V]) Swap(key K, value V) (v V, ok bool) {
	if previous, loaded := this.m.Swap(key, value); loaded {
		if previous != nil {
			v = previous.(V)
		}
		ok = true
	}
	return
}

func (this *Map[K, V]) Get(key K) (v V, b bool) {
	if e, ok := this.m.Load(key); ok {
		if e != nil {
			v = e.(V)
		}
		b = ok
	}
	return
}

func (this *Map[K, V]) Has(key K) (ok bool) {
	_, ok = this.m.Load(key)
	return
}

func (this *Map[K, V]) Del(key K) (ok bool) {
	_, ok = this.m.LoadAndDelete(key)
	return
}

func (this *Map[K, V]) Range(f func(k K, v V) bool) {
	this.m.Range(func(k, v any) bool {
		if v != nil {
			return f(k.(K), v.(V))
		} else {
			var t V
			return f(k.(K), t)
		}
	})
}