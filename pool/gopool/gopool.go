// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/pool/gopool

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

func (fc *funcn) add(f func()) {
	fc.task <- f
	fc.once.Do(func() {
		go fc.run()
	})
}

func (fc *funcn) run() {
	defer func() {
		if err := recover(); err != nil {
			if fc.pool.put(fc) {
				go fc.run()
			}
		}
	}()
	for {
		select {
		case f := <-fc.task:
			f()
		case <-fc.pool.ctx.Done():
			goto END
		}
		if !fc.pool.put(fc) {
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

func (g *GoPool) Go(f func()) {
	if !g.close {
		g.funcnPool <- f
		atomic.AddInt64(&g.tnum, 1)
		if g.mux.TryLock() {
			go g.funcn()
		}
	} else {
		go f()
	}
}

// NumUnExecu the number of functions not executed
func (g *GoPool) NumUnExecu() int {
	return int(g.tnum)
}

// Close when Close ,the pool will enable goroutine, and the func in the pool will be started with goroutine
func (g *GoPool) Close() {
	defer func() { recover() }()
	if atomic.CompareAndSwapInt32(&g._closeflag, 0, 1) {
		g.close = true
		g.cancel()
		go func() {
			<-time.After(10 * time.Second)
			close(g.funcnPool)
		}()
		close(g.pool)
	}
}

func (g *GoPool) TurnOff(off bool) {
	if atomic.CompareAndSwapInt32(&g._closeflag, 0, 0) && off {
		g.close = true
	} else if atomic.CompareAndSwapInt32(&g._closeflag, 0, 0) && !off {
		g.close = false
	}
}

func (g *GoPool) funcn() {
	defer func() { recover() }()
	defer g.mux.Unlock()
	g._funcn()
}

func (g *GoPool) _funcn() {
	for f := range g.funcnPool {
		if g.close {
			go f()
		} else {
			var t *funcn
			if count := atomic.AddInt64(&g.count, 1); count > g.minlimit && count <= g.maxlimit {
				t = &funcn{task: make(chan func(), 1), id: atomic.AddInt64(&g.id, 1), pool: g}
			} else if g.id > g.minlimit {
				t = <-g.pool
			} else if id := atomic.AddInt64(&g.id, 1); id <= g.minlimit {
				t = &funcn{task: make(chan func(), 1), id: id, pool: g}
			} else {
				t = <-g.pool
			}
			t.add(f)
		}
		if atomic.AddInt64(&g.tnum, -1) <= 0 {
			break
		}
	}
	if g.tnum > 0 {
		g._funcn()
	}
	return
}

func (g *GoPool) put(f *funcn) (ok bool) {
	atomic.AddInt64(&g.count, -1)
	if f.id <= g.minlimit && atomic.CompareAndSwapInt32(&g._closeflag, 0, 0) {
		g.pool <- f
		ok = true
	}
	return
}
