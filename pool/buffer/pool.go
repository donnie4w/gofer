// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
// github.com/donnie4w/gofer

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
	_r = p.pool.Get().(*T)
	if p.reset != nil {
		p.reset(_r)
	}
	return
}

func (p *pool[T]) Put(t **T) (ok bool) {
	if *t != nil {
		p.pool.Put(*t)
		*t = nil
		ok = true
	}
	return
}
