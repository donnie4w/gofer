// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/lock

package lock

import (
	"errors"
	"sync"
	"time"

	"github.com/donnie4w/gofer/util"
)

// Numlock provides a set of mutex locks based on integer keys.
type Numlock struct {
	locks  []*sync.Mutex
	muxNum uint64
}

// NewNumLock initializes and returns a new Numlock instance.
func NewNumLock(muxNum int) *Numlock {
	if muxNum <= 0 {
		panic("muxNum must be > 0")
	}
	locks := make([]*sync.Mutex, muxNum)
	for i := range locks {
		locks[i] = &sync.Mutex{}
	}
	return &Numlock{locks: locks, muxNum: uint64(muxNum)}
}

func (nl *Numlock) index(key int64) uint64 {
	return uint64(key) % nl.muxNum
}

// Lock acquires the lock associated with the given key.
func (nl *Numlock) Lock(key int64) *sync.Mutex {
	l := nl.locks[nl.index(key)]
	l.Lock()
	return l
}

// TryLock try to acquire the lock associated with the given key.
func (nl *Numlock) TryLock(key int64) (*sync.Mutex, bool) {
	l := nl.locks[nl.index(key)]
	if l.TryLock() {
		return l, true
	}
	return nil, false
}

// Unlock releases the lock associated with the given key.
func (nl *Numlock) Unlock(key int64) {
	nl.locks[nl.index(key)].Unlock()
}

// Strlock provides a set of read-write mutex locks based on string keys.
type Strlock struct {
	locks  []*sync.RWMutex
	muxNum uint64
}

// NewStrlock initializes and returns a new Strlock instance.
func NewStrlock(muxNum int) *Strlock {
	if muxNum <= 0 {
		panic("muxNum must be > 0")
	}
	locks := make([]*sync.RWMutex, muxNum)
	for i := range locks {
		locks[i] = &sync.RWMutex{}
	}
	return &Strlock{locks: locks, muxNum: uint64(muxNum)}
}

func (sl *Strlock) index(key string) uint64 {
	u := util.Hash64([]byte(key))
	return u % sl.muxNum
}

// Lock acquires the write lock associated with the given key.
func (sl *Strlock) Lock(key string) *sync.RWMutex {
	l := sl.locks[sl.index(key)]
	l.Lock()
	return l
}

// TryLock try to acquire the write lock associated with the given key.
func (sl *Strlock) TryLock(key string) (*sync.RWMutex, bool) {
	l := sl.locks[sl.index(key)]
	if l.TryLock() {
		return l, true
	}
	return nil, false
}

// Unlock releases the write lock associated with the given key.
func (sl *Strlock) Unlock(key string) {
	sl.locks[sl.index(key)].Unlock()
}

// RLock acquires the read lock associated with the given key.
func (sl *Strlock) RLock(key string) *sync.RWMutex {
	l := sl.locks[sl.index(key)]
	l.RLock()
	return l
}

// TryRLock try to acquire the read lock associated with the given key.
func (sl *Strlock) TryRLock(key string) (*sync.RWMutex, bool) {
	l := sl.locks[sl.index(key)]
	if l.TryRLock() {
		return l, true
	}
	return nil, false
}

// RUnlock releases the read lock associated with the given key.
func (sl *Strlock) RUnlock(key string) {
	sl.locks[sl.index(key)].RUnlock()
}

// LimitLock limits the number of concurrent operations and can enforce timeouts.
type LimitLock struct {
	ch chan int
	//count   int64
	//_count  int64
	timeout time.Duration
}

// NewLimitLock initializes and returns a new LimitLock instance.
func NewLimitLock(limit int, timeout time.Duration) *LimitLock {
	return &LimitLock{
		ch:      make(chan int, limit),
		timeout: timeout,
	}
}

// Lock acquires a lock with a timeout.
func (ll *LimitLock) Lock() error {
	timer := time.NewTimer(ll.timeout)
	defer timer.Stop()

	select {
	case ll.ch <- 1:
		//atomic.AddInt64(&ll.count, 1)
		return nil
	case <-timer.C:
		return errors.New("timeout")
	}
}

// Unlock releases a lock.
func (ll *LimitLock) Unlock() {
	<-ll.ch
	//atomic.AddInt64(&ll._count, 1)
}

// Cc returns the current concurrency count.
func (ll *LimitLock) Cc() int {
	return len(ll.ch)
}

// LockCount returns the total number of lock acquisitions.
//func (ll *LimitLock) LockCount() int64 {
//	return atomic.LoadInt64(&ll.count)
//}
