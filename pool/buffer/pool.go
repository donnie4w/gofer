// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/pool/buffer

package buffer

import (
	"sync"
)

type pool[T any] struct {
	pool  sync.Pool
	reset func(*T)
}

func NewPool[T any](constructor func() *T, reset func(*T)) *pool[T] {
	if constructor == nil {
		constructor = func() *T {
			return new(T)
		}
	}
	return &pool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return constructor()
			},
		},
		reset: reset,
	}
}

func (p *pool[T]) Get() (_r *T) {
	if a := p.pool.Get(); a != nil {
		_r = a.(*T)
	}
	if p.reset != nil && _r != nil {
		p.reset(_r)
	}
	return
}

func (p *pool[T]) Put(t **T) bool {
	if *t != nil {
		p.pool.Put(*t)
		*t = nil
		return true
	}
	return false
}
