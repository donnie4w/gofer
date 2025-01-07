// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/lock

package lock

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/donnie4w/gofer/hashmap"
	"github.com/donnie4w/gofer/util"
)

// Numlock provides a set of mutex locks based on integer keys.
type Numlock struct {
	lockm  *hashmap.Map[int64, *sync.Mutex] // Map for storing mutexes indexed by keys
	muxNum int                              // Number of mutexes
}

// NewNumLock initializes and returns a new Numlock instance.
func NewNumLock(muxNum int) *Numlock {
	ml := &Numlock{hashmap.NewMap[int64, *sync.Mutex](), muxNum}
	for i := 0; i < muxNum; i++ {
		ml.lockm.Put(int64(i), &sync.Mutex{}) // Initialize mutexes
	}
	return ml
}

// Lock acquires the lock associated with the given key.
func (nl *Numlock) Lock(key int64) {
	l, _ := nl.lockm.Get(int64(uint64(key) % uint64(nl.muxNum)))
	l.Lock()
}

// TryLock try to acquire the lock associated with the given key.
func (nl *Numlock) TryLock(key int64) bool {
	l, _ := nl.lockm.Get(int64(uint64(key) % uint64(nl.muxNum)))
	return l.TryLock()
}

// Unlock releases the lock associated with the given key.
func (nl *Numlock) Unlock(key int64) {
	l, _ := nl.lockm.Get(int64(uint64(key) % uint64(nl.muxNum)))
	l.Unlock()
}

// Strlock provides a set of read-write mutex locks based on string keys.
type Strlock struct {
	lockm  *hashmap.Map[int64, *sync.RWMutex] // Map for storing read-write mutexes indexed by keys
	muxNum int                                // Number of read-write mutexes
}

// NewStrlock initializes and returns a new Strlock instance.
func NewStrlock(muxNum int) *Strlock {
	ml := &Strlock{hashmap.NewMap[int64, *sync.RWMutex](), muxNum}
	for i := 0; i < muxNum; i++ {
		ml.lockm.Put(int64(i), &sync.RWMutex{}) // Initialize read-write mutexes
	}
	return ml
}

// Lock acquires the write lock associated with the given key.
func (sl *Strlock) Lock(key string) {
	u := util.Hash64([]byte(key)) // Calculate hash of key
	l, _ := sl.lockm.Get(int64(u % uint64(sl.muxNum)))
	l.Lock()
}

// TryLock try to acquire the write lock associated with the given key.
func (sl *Strlock) TryLock(key string) bool {
	u := util.Hash64([]byte(key)) // Calculate hash of key
	l, _ := sl.lockm.Get(int64(u % uint64(sl.muxNum)))
	return l.TryLock()
}

// Unlock releases the write lock associated with the given key.
func (sl *Strlock) Unlock(key string) {
	u := util.Hash64([]byte(key)) // Calculate hash of key
	l, _ := sl.lockm.Get(int64(u % uint64(sl.muxNum)))
	l.Unlock()
}

// RLock acquires the read lock associated with the given key.
func (sl *Strlock) RLock(key string) {
	u := util.Hash64([]byte(key)) // Calculate hash of key
	l, _ := sl.lockm.Get(int64(u % uint64(sl.muxNum)))
	l.RLock()
}

// TryRLock try to acquire the read lock associated with the given key.
func (sl *Strlock) TryRLock(key string) bool {
	u := util.Hash64([]byte(key)) // Calculate hash of key
	l, _ := sl.lockm.Get(int64(u % uint64(sl.muxNum)))
	return l.TryRLock()
}

// RUnlock releases the read lock associated with the given key.
func (sl *Strlock) RUnlock(key string) {
	u := util.Hash64([]byte(key)) // Calculate hash of key
	l, _ := sl.lockm.Get(int64(u % uint64(sl.muxNum)))
	l.RUnlock()
}

// LimitLock limits the number of concurrent operations and can enforce timeouts.
type LimitLock struct {
	ch      chan int      // Channel used for limiting concurrency
	count   int64         // Atomic counter for tracking lock acquisitions
	_count  int64         // Atomic counter for tracking lock releases (unused)
	timeout time.Duration // Timeout duration
}

// NewLimitLock initializes and returns a new LimitLock instance.
func NewLimitLock(limit int, timeout time.Duration) *LimitLock {
	ch := make(chan int, limit) // Create a buffered channel
	return &LimitLock{ch: ch, timeout: timeout}
}

// Lock acquires a lock with a timeout.
func (ll *LimitLock) Lock() (err error) {
	select {
	case <-time.After(ll.timeout): // Wait until timeout
		err = errors.New("timeout")
	case ll.ch <- 1: // Acquire lock
		atomic.AddInt64(&ll.count, 1) // Increment lock count
	}
	return
}

// Unlock releases a lock.
func (ll *LimitLock) Unlock() {
	<-ll.ch                        // Release lock
	atomic.AddInt64(&ll._count, 1) // Increment release count (unused)
}

// Cc returns the current concurrency count.
func (ll *LimitLock) Cc() int64 {
	return ll.count - ll._count // Difference between acquisitions and releases (always zero)
}

// LockCount returns the total number of lock acquisitions.
func (ll *LimitLock) LockCount() int64 {
	return atomic.LoadInt64(&ll.count) // Current lock acquisition count
}

func recoverpanic(err *error) {
	if r := recover(); r != nil {
		if err != nil {
			*err = fmt.Errorf("%v", r)
		}
	}
}
