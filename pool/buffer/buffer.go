// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/pool/buffer

package buffer

import (
	"bytes"
	"sync"
)

type BufferPool struct {
	pool   []sync.Pool
	router []int
}

func NewBufferPool(poolsize int) *BufferPool {
	p := &BufferPool{}
	p.pool = make([]sync.Pool, poolsize)
	p.router = make([]int, poolsize)
	for i := 0; i < poolsize; i++ {
		p.pool[i].New = func() any { return bytes.NewBuffer([]byte{}) }
		p.router[i] = 8 * (1 << i)
	}
	return p
}

func (this *BufferPool) Get(len int) (_r *bytes.Buffer) {
	pre := this.getRouter(len)
	_r = this.pool[pre].Get().(*bytes.Buffer)
	_r.Reset()
	return
}

func (this *BufferPool) Put(buf *bytes.Buffer) (ok bool) {
	if buf != nil {
		pre := this.getRouter(buf.Cap())
		this.pool[pre].Put(buf)
		ok = true
	}
	return true
}

func (this *BufferPool) getRouter(len int) (pre int) {
	for i, v := range this.router {
		if len >= v {
			pre = i
			break
		}
	}
	return
}
