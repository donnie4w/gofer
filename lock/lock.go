// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/tim/lock

package lock

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/donnie4w/gofer/hashmap"
	. "github.com/donnie4w/gofer/util"
)

// ////////////////////////////////////////

type Numlock struct {
	lockm  *Map[int64, *sync.Mutex]
	muxNum int
}

func NewNumLock(muxNum int) *Numlock {
	ml := &Numlock{NewMap[int64, *sync.Mutex](), muxNum}
	for i := 0; i < muxNum; i++ {
		ml.lockm.Put(int64(i), &sync.Mutex{})
	}
	return ml
}

func (this *Numlock) Lock(key int64) {
	l, _ := this.lockm.Get(int64(uint64(key) % uint64(this.muxNum)))
	l.Lock()
}

func (this *Numlock) Unlock(key int64) {
	l, _ := this.lockm.Get(int64(uint64(key) % uint64(this.muxNum)))
	l.Unlock()
}

/*****************************************************/
type Strlock struct {
	lockm  *Map[int64, *sync.RWMutex]
	muxNum int
}

func NewStrlock(muxNum int) *Strlock {
	ml := &Strlock{NewMap[int64, *sync.RWMutex](), muxNum}
	for i := 0; i < muxNum; i++ {
		ml.lockm.Put(int64(i), &sync.RWMutex{})
	}
	return ml
}

func (this *Strlock) Lock(key string) {
	u := Hash([]byte(key))
	l, _ := this.lockm.Get(int64(u % uint64(this.muxNum)))
	l.Lock()
}

func (this *Strlock) Unlock(key string) {
	u := Hash([]byte(key))
	l, _ := this.lockm.Get(int64(u % uint64(this.muxNum)))
	l.Unlock()
}

func (this *Strlock) RLock(key string) {
	u := Hash([]byte(key))
	l, _ := this.lockm.Get(int64(u % uint64(this.muxNum)))
	l.RLock()
}

func (this *Strlock) RUnlock(key string) {
	u := Hash([]byte(key))
	l, _ := this.lockm.Get(int64(u % uint64(this.muxNum)))
	l.RUnlock()
}

/*********************************************************/
type Await[T any] struct {
	m   *Map[int64, chan T]
	mux *Numlock
}

func NewAwait[T any](muxlimit int) *Await[T] {
	return &Await[T]{NewMap[int64, chan T](), NewNumLock(muxlimit)}
}

func (this *Await[T]) Get(idx int64) (ch chan T) {
	defer this.mux.Unlock(idx)
	this.mux.Lock(idx)
	var ok bool
	if ch, ok = this.m.Get(idx); !ok {
		ch = make(chan T, 1)
		this.m.Put(idx, ch)
	}
	return
}

func (this *Await[T]) DelAndClose(idx int64) {
	defer recover()
	if this.m.Has(idx) {
		defer this.mux.Unlock(idx)
		this.mux.Lock(idx)
		if o, ok := this.m.Get(idx); ok {
			close(o)
			this.m.Del(idx)
		}
	}
}

func (this *Await[T]) DelAndPut(idx int64, v T) {
	defer recover()
	if this.m.Has(idx) {
		this.mux.Lock(idx)
		defer this.mux.Unlock(idx)
		if o, ok := this.m.Get(idx); ok {
			o <- v
			this.m.Del(idx)
		}
	}
}

/*********************************************************/
type LimitLock struct {
	ch      chan int
	count   int64
	_count  int64
	timeout time.Duration
}

func NewLimitLock(limit int, timeout time.Duration) (_r *LimitLock) {
	ch := make(chan int, limit)
	_r = &LimitLock{ch: ch, timeout: timeout}
	return
}

func (this *LimitLock) Lock() (err error) {
	select {
	case <-time.After(this.timeout):
		err = errors.New("timeout")
	case this.ch <- 1:
		atomic.AddInt64(&this.count, 1)
	}
	return
}

func (this *LimitLock) Unlock() {
	<-this.ch
	atomic.AddInt64(&this._count, 1)
}

// concurrency num
func (this *LimitLock) Cc() int64 {
	return this.count - this._count
}

func (this *LimitLock) LockCount() int64 {
	return this.count
}
