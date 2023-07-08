// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
// github.com/donnie4w/gofer

package gopool

import (
	"sync"
	"sync/atomic"
)

type funcn struct {
	pool *GoPool
	id   int64
	task chan func()
	once sync.Once
}

func (this *funcn) add(f func()) {
	this.task <- f
	this.once.Do(func() {
		go this.run()
	})
}

func (this *funcn) run() {
	defer func() {
		if err := recover(); err != nil {
			if this.pool.put(this) {
				go this.run()
			}
		}
	}()
	for {
		select {
		case f := <-this.task:
			f()
		}
		if !this.pool.put(this) {
			break
		}
	}
}

type GoPool struct {
	pool      chan *funcn
	minlimit  int64
	maxlimit  int64
	id        int64
	count     int64
	mux       *sync.Mutex
	funcnPool chan func()
}

func NewPool(minlimit int64, maxlimit int64) *GoPool {
	p := &GoPool{}
	p.pool = make(chan *funcn, minlimit)
	p.minlimit, p.maxlimit = minlimit, maxlimit
	if maxlimit < minlimit {
		p.maxlimit = minlimit
	}
	p.funcnPool = make(chan func(), 1<<25)
	p.mux = &sync.Mutex{}
	return p
}

func (this *GoPool) Go(f func()) {
	this.funcnPool <- f
	if this.mux.TryLock() {
		go this.funcn()
	}
}

func (this *GoPool) funcn() {
	defer this.mux.Unlock()
	for len(this.funcnPool) > 0 {
		var t *funcn
		if count := atomic.AddInt64(&this.count, 1); count > this.minlimit && count <= this.maxlimit {
			t = &funcn{task: make(chan func(), 1), id: atomic.AddInt64(&this.id, 1), pool: this}
		} else if this.id > this.minlimit {
			t = <-this.pool
		} else if id := atomic.AddInt64(&this.id, 1); id <= this.minlimit {
			t = &funcn{task: make(chan func(), 1), id: id, pool: this}
		} else {
			t = <-this.pool
		}
		t.add(<-this.funcnPool)
	}
	return
}

func (this *GoPool) put(f *funcn) (ok bool) {
	atomic.AddInt64(&this.count, -1)
	if f.id <= this.minlimit {
		this.pool <- f
		ok = true
	}
	return
}
