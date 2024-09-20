// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/gofer/hashmap

package hashmap

import (
	"sync"
	"sync/atomic"
)

// MapL Define a generic synchronized map structure that also keeps track of its length.
// It uses sync.Map as the underlying data structure and an int64 variable to count the number of elements.
type MapL[K any, V any] struct {
	m   sync.Map
	len int64
}

// NewMapL Create a new instance of a generic MapL and return its pointer.
// The initial length is set to 0.
func NewMapL[K any, V any]() *MapL[K, V] {
	return &MapL[K, V]{m: sync.Map{}}
}

func (t *MapL[K, V]) Put(key K, value V) {
	if _, ok := t.m.Swap(key, value); !ok {
		atomic.AddInt64(&t.len, 1)
	}
}

func (t *MapL[K, V]) Get(key K) (_r V, b bool) {
	if v, ok := t.m.Load(key); ok {
		if v != nil {
			_r = v.(V)
		}
		b = true
	}
	return
}

func (t *MapL[K, V]) Has(key K) (ok bool) {
	_, ok = t.m.Load(key)
	return
}

func (t *MapL[K, V]) Del(key K) (ok bool) {
	if _, ok = t.m.LoadAndDelete(key); ok {
		atomic.AddInt64(&t.len, -1)
	}
	return
}

// Swap the value associated with the given key in the MapL.
// If the key exists, return the old value and true; otherwise, return the zero value and false.
func (t *MapL[K, V]) Swap(key K, value V) (_r V, loaded bool) {
	var previous any
	if previous, loaded = t.m.Swap(key, value); !loaded {
		atomic.AddInt64(&t.len, 1)
	} else if previous != nil {
		_r = previous.(V)
	}
	return
}

// Range  Iterate over all elements in the MapL.
// For each element, call the provided function with the key and value.
// The iteration stops if the function returns false.
func (t *MapL[K, V]) Range(f func(k K, v V) bool) {
	t.m.Range(func(k, v any) bool {
		if v != nil {
			return f(k.(K), v.(V))
		} else {
			var t V
			return f(k.(K), t)
		}
	})
}

func (t *MapL[K, V]) Len() int64 {
	return t.len
}

// Map  Define a generic synchronized map structure using sync.Map as the underlying data structure.
type Map[K any, V any] struct {
	m sync.Map
}

// NewMap Create a new instance of a generic Map and return its pointer.
func NewMap[K any, V any]() *Map[K, V] {
	return &Map[K, V]{m: sync.Map{}}
}

func (t *Map[K, V]) Put(key K, value V) {
	t.m.Swap(key, value)
}

func (t *Map[K, V]) Get(key K) (v V, b bool) {
	if e, ok := t.m.Load(key); ok {
		if e != nil {
			v = e.(V)
		}
		b = ok
	}
	return
}

func (t *Map[K, V]) Has(key K) (ok bool) {
	_, ok = t.m.Load(key)
	return
}

func (t *Map[K, V]) Del(key K) (ok bool) {
	_, ok = t.m.LoadAndDelete(key)
	return
}

// Swap the value associated with the given key in the Map.
// If the key exists, return the old value and true; otherwise, return the zero value and false.
func (t *Map[K, V]) Swap(key K, value V) (_r V, loaded bool) {
	var previous any
	if previous, loaded = t.m.Swap(key, value); loaded && previous != nil {
		_r = previous.(V)
	}
	return
}

// Range  Iterate over all elements in the Map.
// For each element, call the provided function with the key and value.
// The iteration stops if the function returns false.
func (t *Map[K, V]) Range(f func(k K, v V) bool) {
	t.m.Range(func(k, v any) bool {
		if v != nil {
			return f(k.(K), v.(V))
		} else {
			var t V
			return f(k.(K), t)
		}
	})
}

/***********************************************************/
//
//// SortMap the big numbers come front
//type SortMap[K int | int64 | int8 | int32 | string, V any] struct {
//	l   *list.List
//	m   *Map[K, V]
//	mux *sync.RWMutex
//}
//
//func NewSortMap[K int | int64 | int8 | int32 | string, V any]() *SortMap[K, V] {
//	return &SortMap[K, V]{l: list.New(), m: NewMap[K, V](), mux: &sync.RWMutex{}}
//}
//
//func (t *SortMap[K, V]) Put(key K, value V) {
//	defer t.mux.Unlock()
//	t.mux.Lock()
//	t.m.Put(key, value)
//	t.l.PushFront(key)
//	t._swap(t.l.Front())
//}
//
//func (t *SortMap[K, V]) Get(key K) (v V, ok bool) {
//	v, ok = t.m.Get(key)
//	return
//}
//
//func (t *SortMap[K, V]) GetFrontKey() (k K, ok bool) {
//	defer t.mux.RUnlock()
//	t.mux.RLock()
//	if e := t.l.Front(); e != nil {
//		if e.Value != nil {
//			k, ok = e.Value.(K), true
//		} else {
//			ok = true
//		}
//	}
//	return
//}
//
//func (t *SortMap[K, V]) FrontForEach(f func(k K, v V) bool) {
//	defer t.mux.RUnlock()
//	t.mux.RLock()
//	for e := t.l.Front(); e != nil; e = e.Next() {
//		if e.Value != nil {
//			k := e.Value.(K)
//			if v, ok := t.m.Get(k); !ok || !f(k, v) {
//				break
//			}
//		}
//	}
//}
//
//func (t *SortMap[K, V]) BackForEach(f func(k K, v V) bool) {
//	defer t.mux.RUnlock()
//	t.mux.RLock()
//	for e := t.l.Back(); e != nil; e = e.Prev() {
//		if e.Value != nil {
//			k := e.Value.(K)
//			if v, ok := t.m.Get(k); !ok || !f(k, v) {
//				break
//			}
//		}
//	}
//}
//
//func (t *SortMap[K, V]) _swap(e *list.Element) {
//	if e != nil && e.Next() != nil && e.Value.(K) < e.Next().Value.(K) {
//		t.l.MoveAfter(e, e.Next())
//		t._swap(e)
//	}
//}
//
//func (t *SortMap[K, V]) DelAndLoadBack() (k K, v V) {
//	defer t.mux.Unlock()
//	t.mux.Lock()
//	if e := t.l.Back(); e != nil {
//		t.l.Remove(e)
//		if e.Value != nil {
//			k = e.Value.(K)
//			v, _ = t.m.Get(k)
//			t.m.Del(k)
//		}
//	}
//	return
//}
//
//func (t *SortMap[K, V]) Len() int {
//	return t.l.Len()
//}

/************************************************************/

// LinkedMap
// Deprecated
// Use LinkedHashMap instead.
//type LinkedMap[K, V any] struct {
//	l   *list.List
//	m   *Map[K, *list.Element]
//	mux *sync.Mutex
//}

//// NewLinkedMap
//// Deprecated
//// Use NewLinkedHashMap instead.
//func NewLinkedMap[K, V any]() *LinkedMap[K, V] {
//	return &LinkedMap[K, V]{list.New(), NewMap[K, *list.Element](), &sync.Mutex{}}
//}
//
//func (t *LinkedMap[K, V]) Put(k K, v V) {
//	defer t.mux.Unlock()
//	t.mux.Lock()
//	if e, ok := t.m.Swap(k, t.l.PushFront([]any{k, v})); ok {
//		t.l.Remove(e)
//	}
//}
//
//func (t *LinkedMap[K, V]) Get(key K) (v V, ok bool) {
//	defer recover()
//	if e, ok := t.m.Get(key); ok {
//		return e.Value.([]any)[1].(V), ok
//	}
//	return
//}
//
//func (t *LinkedMap[K, V]) Has(key K) (ok bool) {
//	return t.m.Has(key)
//}
//
//func (t *LinkedMap[K, V]) Len() int {
//	return t.l.Len()
//}
//
//func (t *LinkedMap[K, V]) Del(key K) (ok bool) {
//	defer t.mux.Unlock()
//	t.mux.Lock()
//	if e, ok := t.m.Get(key); ok {
//		t.l.Remove(e)
//		t.m.Del(key)
//		return ok
//	}
//	return
//}
//
//func (t *LinkedMap[K, V]) Prev(key K) (k K, v V, _ok bool) {
//	defer t.mux.Unlock()
//	t.mux.Lock()
//	if e, ok := t.m.Get(key); ok {
//		if _v := e.Prev(); _v != nil && _v.Value != nil {
//			k, v = _v.Value.([]any)[0].(K), _v.Value.([]any)[1].(V)
//		}
//		_ok = true
//	} else {
//		_ok = false
//	}
//	return
//}
//
//func (t *LinkedMap[K, V]) Next(key K) (k K, v V) {
//	defer t.mux.Unlock()
//	t.mux.Lock()
//	if e, ok := t.m.Get(key); ok {
//		if _v := e.Next(); _v != nil && _v.Value != nil {
//			k, v = _v.Value.([]any)[0].(K), _v.Value.([]any)[1].(V)
//		}
//	}
//	return
//}
//
//func (t *LinkedMap[K, V]) Back() (k K) {
//	defer t.mux.Unlock()
//	t.mux.Lock()
//	if e := t.l.Back(); e != nil && e.Value != nil {
//		k = e.Value.([]any)[0].(K)
//	}
//	return
//}
//
//func (t *LinkedMap[K, V]) Front() (k K) {
//	defer recover()
//	defer t.mux.Unlock()
//	t.mux.Lock()
//	if e := t.l.Front(); e != nil && e.Value != nil {
//		k = e.Value.([]any)[0].(K)
//	}
//	return
//}
//
//func (t *LinkedMap[K, V]) BackForEach(f func(k K, v V) bool) {
//	defer recover()
//	for e := t.l.Back(); e != nil; e = e.Prev() {
//		if e.Value != nil {
//			es := e.Value.([]any)
//			if !f(es[0].(K), es[1].(V)) {
//				break
//			}
//		}
//	}
//}
//
//func (t *LinkedMap[K, V]) FrontForEach(f func(k K, v V) bool) {
//	defer recover()
//	for e := t.l.Front(); e != nil; e = e.Next() {
//		if e.Value != nil {
//			es := e.Value.([]any)
//			if !f(es[0].(K), es[1].(V)) {
//				break
//			}
//		}
//	}
//}
