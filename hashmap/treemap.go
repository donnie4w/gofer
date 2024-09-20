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

type TreeMap[K cmp.Ordered, V any] struct {
	tree  *btree.BTree
	mutex sync.RWMutex
}

func NewTreeMap[K cmp.Ordered, V any](degree int) *TreeMap[K, V] {
	return &TreeMap[K, V]{tree: btree.New(degree)}
}

func (m *TreeMap[K, V]) Put(key K, value V) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.tree.ReplaceOrInsert(treeItem[K, V]{Key: key, Value: value})
}

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

func (m *TreeMap[K, V]) Delete(key K) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.tree.Delete(treeItem[K, V]{Key: key})
}

func (m *TreeMap[K, V]) Ascend(iter func(K, V) bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	m.tree.Ascend(func(item btree.Item) bool {
		i := item.(treeItem[K, V])
		return iter(i.Key, i.Value)
	})
}

func (m *TreeMap[K, V]) Descend(iter func(K, V) bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	m.tree.Descend(func(item btree.Item) bool {
		i := item.(treeItem[K, V])
		return iter(i.Key, i.Value)
	})
}
