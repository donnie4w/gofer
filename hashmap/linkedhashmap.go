// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/gofer/hashmap

package hashmap

import (
	"container/list"
	"sync"
)

type LinkedHashMap[K, V any] struct {
	m     *sync.Map
	list  *list.List
	limit int64
	count int64
	mu    sync.Mutex
}

func NewLinkedHashMap[K, V any](limit int64) *LinkedHashMap[K, V] {
	return &LinkedHashMap[K, V]{
		m:     &sync.Map{},
		list:  list.New(),
		limit: limit,
		count: 0,
	}
}

type entrykv[K, V any] struct {
	key   K
	value V
}

func (t *LinkedHashMap[K, V]) Put(k K, v V) {
	kv := &entrykv[K, V]{k, v}
	t.mu.Lock()
	defer t.mu.Unlock()
	elem := t.list.PushFront(kv)
	if preelem, ok := t.m.Swap(k, elem); !ok {
		t.count++
		if t.count > t.limit {
			if oldest := t.list.Back(); oldest != nil {
				t.list.Remove(oldest)
				t.m.LoadAndDelete(oldest.Value.(*entrykv[K, V]).key)
			}
			t.count--
		}
	} else if preelem != nil {
		t.list.Remove(preelem.(*list.Element))
	}
}

func (t *LinkedHashMap[K, V]) LoadOrStore(k K, v V) (actual V, loaded bool) {
	kv := &entrykv[K, V]{k, v}
	t.mu.Lock()
	defer t.mu.Unlock()
	elem := t.list.PushFront(kv)
	if preelem, ok := t.m.Swap(k, elem); !ok {
		t.count++
		if t.count > t.limit {
			if oldest := t.list.Back(); oldest != nil {
				t.list.Remove(oldest)
				t.m.LoadAndDelete(oldest.Value.(*entrykv[K, V]).key)
			}
			t.count--
		}
	} else if preelem != nil {
		t.list.Remove(preelem.(*list.Element))
		if v := preelem.(*list.Element).Value; v != nil {
			actual = v.(*entrykv[K, V]).value
			loaded = true
		}
	}
	return
}

func (t *LinkedHashMap[K, V]) Get(k K) (_r V, b bool) {
	if v, ok := t.m.Load(k); ok {
		if v != nil {
			if vv := v.(*list.Element).Value; vv != nil {
				_r = vv.(*entrykv[K, V]).value
			}
		}
		b = ok
	}
	return
}

func (t *LinkedHashMap[K, V]) Delete(k K) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if elem, ok := t.m.Load(k); ok {
		t.list.Remove(elem.(*list.Element))
		t.m.Delete(k)
		t.count--
	}
}

func (t *LinkedHashMap[K, V]) Back() (K, V, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if oldest := t.list.Back(); oldest != nil {
		entry := oldest.Value.(*entrykv[K, V])
		return entry.key, entry.value, true
	}
	var k K
	var v V
	return k, v, false
}

func (t *LinkedHashMap[K, V]) Front() (K, V, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if newest := t.list.Front(); newest != nil {
		entry := newest.Value.(*entrykv[K, V])
		return entry.key, entry.value, true
	}
	var k K
	var v V
	return k, v, false
}

func (t *LinkedHashMap[K, V]) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.m = &sync.Map{}
	t.list.Init()
	t.count = 0
}

func (t *LinkedHashMap[K, V]) Len() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.count
}

func (t *LinkedHashMap[K, V]) MoveToFront(k K) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if elem, ok := t.m.Load(k); ok {
		t.list.MoveToFront(elem.(*list.Element))
	}
}

type LinkedHashMapIterator[K, V any] struct {
	current *list.Element
	mapRef  *LinkedHashMap[K, V]
	front   bool
}

func (it *LinkedHashMapIterator[K, V]) Next() (K, V, bool) {
	it.mapRef.mu.Lock()
	defer it.mapRef.mu.Unlock()

	if it.current != nil {
		entry := it.current.Value.(*entrykv[K, V])
		if it.front {
			it.current = it.current.Next()
		} else {
			it.current = it.current.Prev()
		}
		return entry.key, entry.value, true
	}
	var k K
	var v V
	return k, v, false
}

func (t *LinkedHashMap[K, V]) Iterator(front bool) *LinkedHashMapIterator[K, V] {
	t.mu.Lock()
	defer t.mu.Unlock()
	var current *list.Element
	if front {
		current = t.list.Front()
	} else {
		current = t.list.Back()
	}
	return &LinkedHashMapIterator[K, V]{current: current, mapRef: t, front: front}
}
