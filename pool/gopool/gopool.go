// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
// github.com/donnie4w/gofer

package gopool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
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
		case <-this.pool.ctx.Done():
			goto END
		}
		if !this.pool.put(this) {
			break
		}
	}
END:
}

type GoPool struct {
	pool       chan *funcn
	minlimit   int64
	maxlimit   int64
	id         int64
	count      int64
	mux        *sync.Mutex
	funcnPool  chan func()
	close      bool
	_closeflag int32
	tnum       int64
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewPool(minlimit int64, maxlimit int64) *GoPool {
	return NewPoolWithFuncLimit(minlimit, maxlimit, 1<<17)
}

func NewPoolWithFuncLimit(minlimit int64, maxlimit int64, FuncLimit int) *GoPool {
	p := &GoPool{}
	p.pool = make(chan *funcn, minlimit)
	p.minlimit, p.maxlimit = minlimit, maxlimit
	if maxlimit < minlimit {
		p.maxlimit = minlimit
	}
	p.funcnPool = make(chan func(), FuncLimit)
	p.mux = &sync.Mutex{}
	p.ctx, p.cancel = context.WithCancel(context.Background())
	return p
}

func (this *GoPool) Go(f func()) {
	if !this.close {
		this.funcnPool <- f
		atomic.AddInt64(&this.tnum, 1)
		if this.mux.TryLock() {
			go this.funcn()
		}
	} else {
		go f()
	}
}

// the number of functions not executed
func (this *GoPool) NumUnExecu() int {
	return int(this.tnum)
}

// when Close ,the pool will enable goroutine, and the func in the pool will be started with goroutine
func (this *GoPool) Close() {
	defer recover()
	if atomic.CompareAndSwapInt32(&this._closeflag, 0, 1) {
		this.close = true
		this.cancel()
		go func() {
			<-time.After(5 * time.Second)
			close(this.funcnPool)
		}()
		close(this.pool)
	}
}

func (this *GoPool) funcn() {
	defer recover()
	defer this.mux.Unlock()
	for f := range this.funcnPool {
		if this.close {
			go f()
		} else {
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
			t.add(f)
		}
		if atomic.AddInt64(&this.tnum, -1) <= 0 {
			break
		}
	}
	return
}

func (this *GoPool) put(f *funcn) (ok bool) {
	atomic.AddInt64(&this.count, -1)
	if f.id <= this.minlimit && !this.close {
		this.pool <- f
		ok = true
	}
	return
}
