// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/hashmap

package hashmap

import (
	. "github.com/donnie4w/gofer/buffer"
	"hash/crc64"
	"sort"
	"sync"
)

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

func (t *Consistenthash) Add(keys ...int64) {
	t.mux.Lock()
	defer t.mux.Unlock()
	for _, key := range keys {
		if _, ok := t._m[key]; !ok {
			t._m[key] = 0
			for i := 0; i < t.replicas; i++ {
				buf := NewBufferByPool()
				buf.Write(int64ToBytes(int64(i)))
				buf.Write(int64ToBytes(key))
				h := hash(buf.Bytes())
				buf.Free()
				t.keys = append(t.keys, h)
				t.m.Put(h, key)
			}
		}
	}
	sort.Slice(t.keys, func(i, j int) bool { return t.keys[i] < t.keys[j] })
}

func (t *Consistenthash) Get(value int64) (node int64, ok bool) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	if t.keys == nil {
		return
	}
	keyu64 := hash(int64ToBytes(value))
	idx := sort.Search(len(t.keys), func(i int) bool { return t.keys[i] >= keyu64 })
	if idx >= len(t.keys) {
		idx = 0
	}
	node, ok = t.m.Get(t.keys[idx])
	return
}

func (t *Consistenthash) GetStr(value string) (node int64, ok bool) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	if t.keys == nil {
		return
	}
	keyu64 := hash([]byte(value))
	idx := sort.Search(len(t.keys), func(i int) bool { return t.keys[i] >= keyu64 })
	if idx >= len(t.keys) {
		idx = 0
	}
	node, ok = t.m.Get(t.keys[idx])
	return
}

func (t *Consistenthash) GetNextNodeStr(value string, step int) (nodes []int64, ok bool) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	if t.keys == nil {
		return
	}
	keyu64 := hash([]byte(value))
	idx := sort.Search(len(t.keys), func(i int) bool { return t.keys[i] >= keyu64 })
	if idx >= len(t.keys) {
		idx = 0
	}
	n1, ok := t.m.Get(t.keys[idx])
	m := map[int64]int8{}
	nodes = make([]int64, 0)
	for idx < len(t.keys)-1 {
		idx++
		if node, ok := t.m.Get(t.keys[idx]); ok && node != n1 {
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

func (t *Consistenthash) Del(key int64) {
	t.mux.Lock()
	defer t.mux.Unlock()
	if _, ok := t._m[key]; ok {
		t.keys = t.keys[:0]
		t.m.Range(func(k uint64, v int64) bool {
			if v == key {
				t.m.Del(k)
			} else {
				t.keys = append(t.keys, k)
			}
			return true
		})
		sort.Slice(t.keys, func(i, j int) bool { return t.keys[i] < t.keys[j] })
		delete(t._m, key)
	}
}

func (t *Consistenthash) Nodes() (_r []int64) {
	t.mux.Lock()
	defer t.mux.Unlock()
	_r = make([]int64, 0)
	for k := range t._m {
		_r = append(_r, k)
	}
	return
}
