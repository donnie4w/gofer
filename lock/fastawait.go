// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/lock

package lock

import (
	"context"
	"fmt"
	"github.com/donnie4w/gofer/cache"
	"github.com/donnie4w/gofer/hashmap"
	"github.com/donnie4w/gofer/util"
	"time"
)

type FastAwait[T any] struct {
	db *hashmap.Map[int64, chan T]
	bf *cache.Bloomfilter
}

func NewFastAwait[T any]() *FastAwait[T] {
	return &FastAwait[T]{db: hashmap.NewMap[int64, chan T](), bf: cache.NewBloomFilter(1 << 20)}
}

func (fa *FastAwait[T]) del(syncId int64) {
	fa.bf.Add(util.Int64ToBytes(syncId))
	fa.db.Del(syncId)
}

func (fa *FastAwait[T]) isDel(syncId int64) bool {
	return fa.bf.Contains(util.Int64ToBytes(syncId))
}

func (fa *FastAwait[T]) Wait(syncId int64, timeout time.Duration) (r T, err error) {
	defer recoverpanic(&err)
	ch := make(chan T, 1)
	fa.db.Put(syncId, ch)
	defer fa.del(syncId)
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		select {
		case <-timer.C:
			defer func() {
				defer recoverpanic(nil)
				close(ch)
			}()
			return r, fmt.Errorf("wait %d timeout", syncId)
		case r = <-ch:
			return
		}
	} else {
		r = <-ch
	}
	return
}

func (fa *FastAwait[T]) WaitWithCancel(ctx context.Context, syncId int64, timeout time.Duration) (r T, cancel bool, err error) {
	defer recoverpanic(&err)
	ch := make(chan T, 1)
	fa.db.Put(syncId, ch)
	defer fa.del(syncId)
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			defer func() {
				defer recoverpanic(nil)
				close(ch)
			}()
			return r, true, fmt.Errorf("wait %d cancel", syncId)
		case <-timer.C:
			defer func() {
				defer recoverpanic(nil)
				close(ch)
			}()
			return r, false, fmt.Errorf("wait %d timeout", syncId)
		case r = <-ch:
			return
		}
	} else {
		r = <-ch
	}
	return
}

func (fa *FastAwait[T]) CloseAndPut(syncId int64, v T) (err error) {
	defer recoverpanic(&err)
	loop := 30
START:
	if ch, ok := fa.db.Get(syncId); ok {
		defer close(ch)
		fa.db.Del(syncId)
		ch <- v
	} else if !fa.isDel(syncId) && loop > 0 {
		loop--
		<-time.After(100 * time.Millisecond)
		goto START
	}
	return
}

func (fa *FastAwait[T]) Close(syncId int64) (err error) {
	defer recoverpanic(&err)
	loop := 30
START:
	if ch, ok := fa.db.Get(syncId); ok {
		defer close(ch)
		fa.db.Del(syncId)
	} else if !fa.isDel(syncId) && loop > 0 {
		loop--
		<-time.After(100 * time.Millisecond)
		goto START
	}
	return
}
