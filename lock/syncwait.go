// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/lock

package lock

import (
	"github.com/donnie4w/gofer/hashmap"
	"sync"
	"time"
)

// SyncWait is a generic wait group implementation that allows setting channels for different keys.
type SyncWait struct {
	m   *hashmap.Map[int64, *sync.WaitGroup] // Map for storing channels indexed by keys
	mux *Numlock                             // Mutex lock for synchronization
}

// NewSyncWait initializes and returns a new Await instance.
func NewSyncWait(muxlimit int) *SyncWait {
	return &SyncWait{hashmap.NewMap[int64, *sync.WaitGroup](), NewNumLock(muxlimit)}
}

// get retrieves or creates a channel associated with the given index.
func (sw *SyncWait) get(idx int64) (wg *sync.WaitGroup) {
	sw.mux.Lock(idx)
	defer sw.mux.Unlock(idx)
	var ok bool
	if wg, ok = sw.m.Get(idx); !ok {
		wg = &sync.WaitGroup{}
		sw.m.Put(idx, wg)
	}
	return
}

// Has checks if a channel exists for the given index.
func (sw *SyncWait) Has(idx int64) bool {
	return sw.m.Has(idx)
}

// Close deletes and closes the channel associated with the given index.
func (sw *SyncWait) Close(idx int64) {
	defer recoverpanic(nil)
	loop := 1000
START:
	if sw.m.Has(idx) {
		sw.mux.Lock(idx)
		defer sw.mux.Unlock(idx)
		if wg, ok := sw.m.Get(idx); ok {
			sw.m.Del(idx)
			wg.Done()
		}
	} else if loop > 0 {
		loop--
		<-time.After(time.Millisecond)
		goto START
	}
}

// Wait wait for the channel data to return or close
func (sw *SyncWait) Wait(idx int64) {
	defer sw.m.Del(idx)
	wg := sw.get(idx)
	wg.Add(1)
	wg.Wait()
}

// WaitWithTimeOut wait for the channel data to return or close, and set the timeout period
//func (sw *SyncWait) WaitWithTimeOut(idx int64, timeout time.Duration) error {
//	defer sw.m.Del(idx)
//	wg := sw.get(idx)
//	ch := make(chan struct{})
//	go func() {
//		wg.Wait()
//		close(ch)
//	}()
//	timer := time.NewTimer(timeout)
//	defer timer.Stop()
//	select {
//	case <-ch:
//		return nil
//	case <-timer.C:
//		return fmt.Errorf("wait %d timeout", idx)
//	}
//}
