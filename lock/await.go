//// Copyright (c) 2023, donnie <donnie4w@gmail.com>
//// All rights reserved.
//// Use of t source code is governed by a BSD-style
//// license that can be found in the LICENSE file.
////
//// github.com/donnie4w/gofer/lock
//
//package lock
//
//import (
//	"context"
//	"fmt"
//	"github.com/donnie4w/gofer/hashmap"
//	"sync"
//	"time"
//)
//
//// Await is a generic wait group implementation that allows setting channels for different keys.
//type Await[T any] struct {
//	m   *hashmap.Map[int64, chan T]
//	mux *Numlock
//}
//
//// NewAwait initializes and returns a new Await instance.
//func NewAwait[T any](muxlimit int) *Await[T] {
//	return &Await[T]{
//		m:   hashmap.NewMap[int64, chan T](),
//		mux: NewNumLock(muxlimit),
//	}
//}
//
//// Get retrieves or creates a channel associated with the given index.
//func (at *Await[T]) Get(idx int64) (ch chan T) {
//	at.mux.Lock(idx)
//	defer at.mux.Unlock(idx)
//
//	if ch, ok := at.m.Get(idx); ok {
//		return ch
//	}
//
//	ch = make(chan T, 1)
//	at.m.Put(idx, ch)
//	return
//}
//
//// Has checks if a channel exists for the given index.
//func (at *Await[T]) Has(idx int64) bool {
//	return at.m.Has(idx)
//}
//
//// Close deletes and closes the channel associated with the given index.
//func (at *Await[T]) Close(idx int64) (err error) {
//	defer recoverfunc(&err)
//
//	for i := 0; i < 1000; i++ {
//
//		at.mux.Lock(idx)
//
//		if ch, ok := at.m.Get(idx); ok {
//			at.m.Del(idx)
//			at.mux.Unlock(idx)
//
//			close(ch)
//			return nil
//		}
//
//		at.mux.Unlock(idx)
//		time.Sleep(time.Millisecond)
//	}
//
//	return nil
//}
//
//// CloseAndPut sends a value to the channel and deletes it from the map.
//func (at *Await[T]) CloseAndPut(idx int64, v T) (err error) {
//	defer recoverfunc(&err)
//
//	for i := 0; i < 1000; i++ {
//
//		at.mux.Lock(idx)
//
//		if ch, ok := at.m.Get(idx); ok {
//			at.m.Del(idx)
//			at.mux.Unlock(idx)
//
//			ch <- v
//			close(ch)
//			return nil
//		}
//
//		at.mux.Unlock(idx)
//		time.Sleep(time.Millisecond)
//	}
//
//	return nil
//}
//
//// WaitWithCancel wait for the channel data to return or close, and set the timeout period
//func (at *Await[T]) WaitWithCancel(ctx context.Context, idx int64, timeout time.Duration) (r T, isCancel bool, err error) {
//	defer recoverfunc(&err)
//
//	ch := at.Get(idx)
//
//	defer at.m.Del(idx)
//
//	if timeout > 0 {
//
//		timer := time.NewTimer(timeout)
//		defer timer.Stop()
//
//		select {
//
//		case <-ctx.Done():
//
//			defer func() {
//				defer recoverfunc(nil)
//				close(ch)
//			}()
//
//			return r, true, fmt.Errorf("cancel %d", idx)
//
//		case <-timer.C:
//
//			defer func() {
//				defer recoverfunc(nil)
//				close(ch)
//			}()
//
//			return r, false, fmt.Errorf("wait %d timeout", idx)
//
//		case r = <-ch:
//			return r, false, nil
//		}
//
//	} else {
//
//		select {
//
//		case <-ctx.Done():
//
//			defer func() {
//				defer recoverfunc(nil)
//				close(ch)
//			}()
//
//			return r, true, fmt.Errorf("cancel %d", idx)
//
//		case r = <-ch:
//			return r, false, nil
//		}
//	}
//}
//
//// Wait wait for the channel data to return or close, and set the timeout period
//func (at *Await[T]) Wait(idx int64, timeout time.Duration) (r T, err error) {
//	defer recoverfunc(&err)
//
//	ch := at.Get(idx)
//
//	defer at.m.Del(idx)
//
//	if timeout > 0 {
//
//		timer := time.NewTimer(timeout)
//		defer timer.Stop()
//
//		select {
//
//		case <-timer.C:
//
//			defer func() {
//				defer recoverfunc(nil)
//				close(ch)
//			}()
//
//			return r, fmt.Errorf("wait %d timeout", idx)
//
//		case r = <-ch:
//			return r, nil
//		}
//
//	} else {
//
//		r = <-ch
//		return r, nil
//	}
//}
//
//// SyncWait wait for the channel data to return or close, and set the timeout period
//func (at *Await[T]) SyncWait(wg *sync.WaitGroup, idx int64, timeout time.Duration) (r T, err error) {
//	defer recoverfunc(&err)
//
//	ch := at.Get(idx)
//
//	defer at.m.Del(idx)
//
//	if wg != nil {
//		wg.Done()
//	}
//
//	if timeout > 0 {
//
//		timer := time.NewTimer(timeout)
//		defer timer.Stop()
//
//		select {
//
//		case <-timer.C:
//
//			defer func() {
//				defer recoverfunc(nil)
//				close(ch)
//			}()
//
//			return r, fmt.Errorf("wait %d timeout", idx)
//
//		case r = <-ch:
//			return r, nil
//		}
//
//	} else {
//
//		r = <-ch
//		return r, nil
//	}
//}
//
//func recoverfunc(err *error) {
//	if r := recover(); r != nil {
//		if err != nil {
//			*err = fmt.Errorf("%v", r)
//		}
//	}
//}

package lock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/donnie4w/gofer/hashmap"
)

const (
	defaultRetryLoop = 30
	retryInterval    = 100 * time.Millisecond
)

type Await[T any] struct {
	m *hashmap.Map[int64, *future[T]]
}

func NewAwait[T any]() *Await[T] {
	return &Await[T]{
		m: hashmap.NewMap[int64, *future[T]](),
	}
}

// getFuture 内部方法（完全复用原 Get 的加锁创建逻辑）
func (at *Await[T]) getFuture(idx int64) *future[T] {
	if v, ok := at.m.Get(idx); ok {
		return v
	}
	f := &future[T]{ch: make(chan T, 1)}
	v, _ := at.m.GetOrInsert(idx, f)
	return v
}

// Get retrieves or creates a channel associated with the given index.
func (at *Await[T]) Get(idx int64) (ch chan T) {
	return at.getFuture(idx).ch
}

// Has checks if a channel exists for the given index.
func (at *Await[T]) Has(idx int64) bool {
	return at.m.Has(idx)
}

func (at *Await[T]) notify(idx int64, v T, withValue bool) error {
	loop := defaultRetryLoop
	for loop > 0 {
		if f, ok := at.m.Get(idx); ok {
			at.m.Del(idx)
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

func (at *Await[T]) Close(idx int64) (err error) {
	defer recoverfunc(&err)
	return at.notify(idx, *new(T), false)
}

func (at *Await[T]) CloseAndPut(idx int64, v T) (err error) {
	defer recoverfunc(&err)
	return at.notify(idx, v, true)
}

// WaitWithCancel wait for the channel data to return or close, and set the timeout period
func (at *Await[T]) WaitWithCancel(ctx context.Context, idx int64, timeout time.Duration) (r T, isCancel bool, err error) {
	defer recoverfunc(&err)

	f := at.getFuture(idx)
	defer at.m.Del(idx)

	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			f.close()
			return r, true, fmt.Errorf("cancel %d", idx)
		case <-timer.C:
			f.close()
			return r, false, fmt.Errorf("wait %d timeout", idx)
		case r = <-f.ch:
			return r, false, nil
		}
	}

	// timeout=0 支持 ctx 取消
	select {
	case <-ctx.Done():
		f.close()
		return r, true, fmt.Errorf("cancel %d", idx)
	case r = <-f.ch:
		return r, false, nil
	}
}

// Wait wait for the channel data to return or close, and set the timeout period
func (at *Await[T]) Wait(idx int64, timeout time.Duration) (r T, err error) {
	defer recoverfunc(&err)

	f := at.getFuture(idx)
	defer at.m.Del(idx)

	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case <-timer.C:
			f.close()
			return r, fmt.Errorf("wait %d timeout", idx)
		case r = <-f.ch:
			return r, nil
		}
	}
	r = <-f.ch
	return r, nil
}

// SyncWait wait for the channel data to return or close, and set the timeout period
func (at *Await[T]) SyncWait(wg *sync.WaitGroup, idx int64, timeout time.Duration) (r T, err error) {
	defer recoverfunc(&err)

	f := at.getFuture(idx)
	defer at.m.Del(idx)

	if timeout > 0 {
		if wg != nil {
			wg.Done()
		}
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case <-timer.C:
			f.close()
			return r, fmt.Errorf("wait %d timeout", idx)
		case r = <-f.ch:
			return r, nil
		}
	}

	if wg != nil {
		wg.Done()
	}
	r = <-f.ch
	return r, nil
}

func recoverfunc(err *error) {
	if r := recover(); r != nil {
		if err != nil {
			*err = fmt.Errorf("%v", r)
		}
	}
}
