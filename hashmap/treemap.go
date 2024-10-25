// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/hashmap

package hashmap

import (
	"cmp"
	"github.com/google/btree"
	"sync"
)

type treeItem[K cmp.Ordered, V any] struct {
	Key   K
	Value V
}

func (a treeItem[K, V]) Less(b btree.Item) bool {
	return a.Key < b.(treeItem[K, V]).Key
}

// TreeMap Define the TreeMap structure which internally uses a B-tree for storing data with an RWMutex for concurrent access control.
type TreeMap[K cmp.Ordered, V any] struct {
	tree  *btree.BTree
	mutex sync.RWMutex
}

// NewTreeMap Create a new instance of TreeMap with a specified degree for the underlying B-tree.
func NewTreeMap[K cmp.Ordered, V any](degree int) *TreeMap[K, V] {
	return &TreeMap[K, V]{tree: btree.New(degree)}
}

// Put Insert or update a key-value pair in the TreeMap.
func (m *TreeMap[K, V]) Put(key K, value V) (prev V, b bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if item := m.tree.ReplaceOrInsert(treeItem[K, V]{Key: key, Value: value}); item != nil {
		prev, b = item.(treeItem[K, V]).Value, true
	}
	return
}

// Get Retrieve a value from the TreeMap by its key, returning the value and a boolean indicating if the key was found.
func (m *TreeMap[K, V]) Get(key K) (V, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	item := m.tree.Get(treeItem[K, V]{Key: key})
	if item != nil {
		return item.(treeItem[K, V]).Value, true
	}
	var zero V
	return zero, false
}

// Del Delete a key from the TreeMap.
func (m *TreeMap[K, V]) Del(key K) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.tree.Delete(treeItem[K, V]{Key: key})
}

// Ascend Traverse all elements in ascending order based on their keys, executing the iterator function.
func (m *TreeMap[K, V]) Ascend(iter func(K, V) bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	m.tree.Ascend(func(item btree.Item) bool {
		i := item.(treeItem[K, V])
		return iter(i.Key, i.Value)
	})
}

// Descend Traverse all elements in descending order based on their keys, executing the iterator function.
func (m *TreeMap[K, V]) Descend(iter func(K, V) bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	m.tree.Descend(func(item btree.Item) bool {
		i := item.(treeItem[K, V])
		return iter(i.Key, i.Value)
	})
}
