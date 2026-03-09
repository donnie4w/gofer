// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/lock

package lock

import (
	"context"
	"fmt"
	"github.com/donnie4w/gofer/hashmap"
	"sync"
	"time"
)

type future[T any] struct {
	ch   chan T
	once sync.Once
}

func (f *future[T]) close() {
	f.once.Do(func() { close(f.ch) })
}

func (f *future[T]) put(v T) {
	f.once.Do(func() {
		f.ch <- v
		close(f.ch)
	})
}

type FastAwait[T any] struct {
	db *hashmap.Map[int64, *future[T]]
}

func NewFastAwait[T any]() *FastAwait[T] {
	return &FastAwait[T]{db: hashmap.NewMap[int64, *future[T]]()}
}

func (fa *FastAwait[T]) getFuture(syncId int64) *future[T] {
	if v, ok := fa.db.Get(syncId); ok {
		return v
	}
	f := &future[T]{ch: make(chan T, 1)}
	v, _ := fa.db.GetOrInsert(syncId, f)
	return v
}

func (fa *FastAwait[T]) Wait(syncId int64, timeout time.Duration) (r T, err error) {
	defer recoverfunc(&err)

	f := fa.getFuture(syncId)
	defer fa.db.Del(syncId)

	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case <-timer.C:
			f.close()
			return r, fmt.Errorf("wait %d timeout", syncId)
		case r = <-f.ch:
			return
		}
	}
	r = <-f.ch
	return
}

func (fa *FastAwait[T]) WaitWithCancel(ctx context.Context, syncId int64, timeout time.Duration) (r T, cancel bool, err error) {
	defer recoverfunc(&err)

	f := fa.getFuture(syncId)
	defer fa.db.Del(syncId)

	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			f.close()
			return r, true, fmt.Errorf("wait %d cancel", syncId)
		case <-timer.C:
			f.close()
			return r, false, fmt.Errorf("wait %d timeout", syncId)
		case r = <-f.ch:
			return
		}
	}

	select {
	case <-ctx.Done():
		f.close()
		return r, true, fmt.Errorf("wait %d cancel", syncId)
	case r = <-f.ch:
		return
	}
}

func (fa *FastAwait[T]) CloseAndPut(syncId int64, v T) (err error) {
	defer recoverfunc(&err)
	return fa.notify(syncId, v, true)
}

func (fa *FastAwait[T]) Close(syncId int64) (err error) {
	defer recoverfunc(&err)
	return fa.notify(syncId, *new(T), false)
}

func (fa *FastAwait[T]) notify(syncId int64, v T, withValue bool) error {
	loop := defaultRetryLoop
	for loop > 0 {
		if f, ok := fa.db.Get(syncId); ok {
			fa.db.Del(syncId)
			if withValue {
				f.put(v)
			} else {
				f.close()
			}
			return nil
		}
		loop--
		time.Sleep(retryInterval)
	}
	return nil // If registration is not completed within 3 seconds, the data will be discarded.
}
