// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/gofer/hashmap

package hashmap

import (
	"container/list"
	"hash/crc64"
	"sort"
	"sync"
	"sync/atomic"

	. "github.com/donnie4w/gofer/buffer"
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

/***********************************************************/

// the big numbers come front
type SortMap[K int | int64 | int8 | int32 | string, V any] struct {
	l   *list.List
	m   *Map[K, V]
	mux *sync.RWMutex
}

func NewSortMap[K int | int64 | int8 | int32 | string, V any]() *SortMap[K, V] {
	return &SortMap[K, V]{l: list.New(), m: NewMap[K, V](), mux: &sync.RWMutex{}}
}

func (this *SortMap[K, V]) Put(key K, value V) {
	defer this.mux.Unlock()
	this.mux.Lock()
	this.m.Put(key, value)
	this.l.PushFront(key)
	this._swap(this.l.Front())
}

func (this *SortMap[K, V]) Get(key K) (v V, ok bool) {
	v, ok = this.m.Get(key)
	return
}

func (this *SortMap[K, V]) GetFrontKey() (k K, ok bool) {
	defer this.mux.RUnlock()
	this.mux.RLock()
	if e := this.l.Front(); e != nil {
		if e.Value != nil {
			k, ok = e.Value.(K), true
		} else {
			ok = true
		}
	}
	return
}

func (this *SortMap[K, V]) FrontForEach(f func(k K, v V) bool) {
	defer this.mux.RUnlock()
	this.mux.RLock()
	for e := this.l.Front(); e != nil; e = e.Next() {
		if e.Value != nil {
			k := e.Value.(K)
			if v, ok := this.m.Get(k); !ok || !f(k, v) {
				break
			}
		}
	}
}

func (this *SortMap[K, V]) BackForEach(f func(k K, v V) bool) {
	defer this.mux.RUnlock()
	this.mux.RLock()
	for e := this.l.Back(); e != nil; e = e.Prev() {
		if e.Value != nil {
			k := e.Value.(K)
			if v, ok := this.m.Get(k); !ok || !f(k, v) {
				break
			}
		}
	}
}

func (this *SortMap[K, V]) _swap(e *list.Element) {
	if e != nil && e.Next() != nil && e.Value.(K) < e.Next().Value.(K) {
		this.l.MoveAfter(e, e.Next())
		this._swap(e)
	}
}

func (this *SortMap[K, V]) DelAndLoadBack() (k K, v V) {
	defer this.mux.Unlock()
	this.mux.Lock()
	if e := this.l.Back(); e != nil {
		this.l.Remove(e)
		if e.Value != nil {
			k = e.Value.(K)
			v, _ = this.m.Get(k)
			this.m.Del(k)
		}
	}
	return
}

func (this *SortMap[K, V]) Len() int {
	return this.l.Len()
}

/************************************************************/
type LinkedMap[K, V any] struct {
	l   *list.List
	m   *Map[K, *list.Element]
	mux *sync.Mutex
}

func NewLinkedMap[K, V any]() *LinkedMap[K, V] {
	return &LinkedMap[K, V]{list.New(), NewMap[K, *list.Element](), &sync.Mutex{}}
}

func (this *LinkedMap[K, V]) Put(k K, v V) {
	defer this.mux.Unlock()
	this.mux.Lock()
	if e, ok := this.m.Swap(k, this.l.PushFront([]any{k, v})); ok {
		this.l.Remove(e)
	}
}

func (this *LinkedMap[K, V]) Get(key K) (v V, ok bool) {
	defer recover()
	if e, ok := this.m.Get(key); ok {
		return e.Value.([]any)[1].(V), ok
	}
	return
}

func (this *LinkedMap[K, V]) Has(key K) (ok bool) {
	return this.m.Has(key)
}

func (this *LinkedMap[K, V]) Len() int {
	return this.l.Len()
}

func (this *LinkedMap[K, V]) Del(key K) (ok bool) {
	defer this.mux.Unlock()
	this.mux.Lock()
	if e, ok := this.m.Get(key); ok {
		this.l.Remove(e)
		this.m.Del(key)
		return ok
	}
	return
}

func (this *LinkedMap[K, V]) Prev(key K) (k K, v V, _ok bool) {
	defer this.mux.Unlock()
	this.mux.Lock()
	if e, ok := this.m.Get(key); ok {
		if _v := e.Prev(); _v != nil && _v.Value != nil {
			k, v = _v.Value.([]any)[0].(K), _v.Value.([]any)[1].(V)
		}
		_ok = true
	} else {
		_ok = false
	}
	return
}

func (this *LinkedMap[K, V]) Next(key K) (k K, v V) {
	defer this.mux.Unlock()
	this.mux.Lock()
	if e, ok := this.m.Get(key); ok {
		if _v := e.Next(); _v != nil && _v.Value != nil {
			k, v = _v.Value.([]any)[0].(K), _v.Value.([]any)[1].(V)
		}
	}
	return
}

func (this *LinkedMap[K, V]) Back() (k K) {
	defer this.mux.Unlock()
	this.mux.Lock()
	if e := this.l.Back(); e != nil && e.Value != nil {
		k = e.Value.([]any)[0].(K)
	}
	return
}

func (this *LinkedMap[K, V]) Front() (k K) {
	defer recover()
	defer this.mux.Unlock()
	this.mux.Lock()
	if e := this.l.Front(); e != nil && e.Value != nil {
		k = e.Value.([]any)[0].(K)
	}
	return
}

func (this *LinkedMap[K, V]) BackForEach(f func(k K, v V) bool) {
	defer recover()
	for e := this.l.Back(); e != nil; e = e.Prev() {
		if e.Value != nil {
			es := e.Value.([]any)
			if !f(es[0].(K), es[1].(V)) {
				break
			}
		}
	}
}

func (this *LinkedMap[K, V]) FrontForEach(f func(k K, v V) bool) {
	defer recover()
	for e := this.l.Front(); e != nil; e = e.Next() {
		if e.Value != nil {
			es := e.Value.([]any)
			if !f(es[0].(K), es[1].(V)) {
				break
			}
		}
	}
}

type LimitMap[K, V any] struct {
	m     *sync.Map
	ch    chan K
	limit int64
	count int64
}

func NewLimitMap[K, V any](limit int64) *LimitMap[K, V] {
	return &LimitMap[K, V]{&sync.Map{}, make(chan K, limit), limit, 0}
}

func (this *LimitMap[K, V]) Put(k K, v V) {
	if _, ok := this.m.Swap(k, v); !ok {
		if atomic.AddInt64(&this.count, 1) >= this.limit {
			if _, ok := this.m.LoadAndDelete(<-this.ch); ok {
				atomic.AddInt64(&this.count, -1)
			}
		}
		this.ch <- k
	}
}

func (this *LimitMap[K, V]) LoadOrStore(k K, v V) (actual V, loaded bool) {
	if previous, ok := this.m.LoadOrStore(k, v); !ok {
		if atomic.AddInt64(&this.count, 1) >= this.limit {
			if _, ok := this.m.LoadAndDelete(<-this.ch); ok {
				atomic.AddInt64(&this.count, -1)
			}
		}
		this.ch <- k
	} else {
		if previous != nil {
			actual = previous.(V)
		}
		loaded = ok
	}
	return
}

func (this *LimitMap[K, V]) Get(k K) (_r V, b bool) {
	if v, ok := this.m.Load(k); ok {
		if v != nil {
			_r = v.(V)
		}
		b = ok
	}
	return
}

func (this *LimitMap[K, V]) Has(k K) (b bool) {
	_, b = this.m.Load(k)
	return
}

/***********************************************************/

func int64ToBytes(n int64) (bs []byte) {
	bs = make([]byte, 8)
	for i := 0; i < 8; i++ {
		bs[i] = byte(n >> (8 * (7 - i)))
	}
	return
}

func hash(bs []byte) uint64 {
	return crc64.Checksum(bs, crc64.MakeTable(crc64.ECMA))
}

type Consistenthash struct {
	replicas int
	keys     []uint64
	m        *Map[uint64, int64]
	mux      *sync.RWMutex
	_m       map[int64]int8
}

func NewConsistenthash(replicas int) (m *Consistenthash) {
	m = &Consistenthash{
		replicas: replicas,
		m:        NewMap[uint64, int64](),
		mux:      &sync.RWMutex{},
		_m:       map[int64]int8{},
	}
	return m
}

func (this *Consistenthash) Add(keys ...int64) {
	this.mux.Lock()
	defer this.mux.Unlock()
	for _, key := range keys {
		if _, ok := this._m[key]; !ok {
			this._m[key] = 0
			for i := 0; i < this.replicas; i++ {
				buf := NewBufferByPool()
				buf.Write(int64ToBytes(int64(i)))
				buf.Write(int64ToBytes(key))
				h := hash(buf.Bytes())
				buf.Free()
				this.keys = append(this.keys, h)
				this.m.Put(h, key)
			}
		}
	}
	sort.Slice(this.keys, func(i, j int) bool { return this.keys[i] < this.keys[j] })
}

func (this *Consistenthash) Get(value int64) (node int64, ok bool) {
	this.mux.RLock()
	defer this.mux.RUnlock()
	if this.keys == nil {
		return
	}
	keyu64 := hash(int64ToBytes(value))
	idx := sort.Search(len(this.keys), func(i int) bool { return this.keys[i] >= keyu64 })
	if idx >= len(this.keys) {
		idx = 0
	}
	node, ok = this.m.Get(this.keys[idx])
	return
}

func (this *Consistenthash) GetStr(value string) (node int64, ok bool) {
	this.mux.RLock()
	defer this.mux.RUnlock()
	if this.keys == nil {
		return
	}
	keyu64 := hash([]byte(value))
	idx := sort.Search(len(this.keys), func(i int) bool { return this.keys[i] >= keyu64 })
	if idx >= len(this.keys) {
		idx = 0
	}
	node, ok = this.m.Get(this.keys[idx])
	return
}

func (this *Consistenthash) GetNextNodeStr(value string, step int) (nodes []int64, ok bool) {
	this.mux.RLock()
	defer this.mux.RUnlock()
	if this.keys == nil {
		return
	}
	keyu64 := hash([]byte(value))
	idx := sort.Search(len(this.keys), func(i int) bool { return this.keys[i] >= keyu64 })
	if idx >= len(this.keys) {
		idx = 0
	}
	n1, ok := this.m.Get(this.keys[idx])
	m := map[int64]int8{}
	nodes = make([]int64, 0)
	for idx < len(this.keys)-1 {
		idx++
		if node, ok := this.m.Get(this.keys[idx]); ok && node != n1 {
			if _, ok := m[node]; !ok {
				m[node] = 0
				nodes = append(nodes, node)
			}
		}
		if len(nodes) == step {
			return
		}
	}
	return
}

func (this *Consistenthash) Del(key int64) {
	this.mux.Lock()
	defer this.mux.Unlock()
	if _, ok := this._m[key]; ok {
		this.keys = this.keys[:0]
		this.m.Range(func(k uint64, v int64) bool {
			if v == key {
				this.m.Del(k)
			} else {
				this.keys = append(this.keys, k)
			}
			return true
		})
		sort.Slice(this.keys, func(i, j int) bool { return this.keys[i] < this.keys[j] })
		delete(this._m, key)
	}
}

func (this *Consistenthash) Nodes() (_r []int64) {
	this.mux.Lock()
	defer this.mux.Unlock()
	_r = make([]int64, 0)
	for k := range this._m {
		_r = append(_r, k)
	}
	return
}
