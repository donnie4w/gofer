// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/lock

package lock

import (
	"fmt"
	"github.com/donnie4w/gofer/hashmap"
	"time"
)

// Await is a generic wait group implementation that allows setting channels for different keys.
type Await[T any] struct {
	m   *hashmap.Map[int64, chan T] // Map for storing channels indexed by keys
	mux *Numlock                    // Mutex lock for synchronization
}

// NewAwait initializes and returns a new Await instance.
func NewAwait[T any](muxlimit int) *Await[T] {
	return &Await[T]{hashmap.NewMap[int64, chan T](), NewNumLock(muxlimit)}
}

// Get retrieves or creates a channel associated with the given index.
func (at *Await[T]) Get(idx int64) (ch chan T) {
	at.mux.Lock(idx)
	defer at.mux.Unlock(idx)
	var ok bool
	if ch, ok = at.m.Get(idx); !ok {
		ch = make(chan T, 1) // Create new channel if not found
		at.m.Put(idx, ch)
	}
	return
}

// Has checks if a channel exists for the given index.
func (at *Await[T]) Has(idx int64) bool {
	return at.m.Has(idx)
}

// Close deletes and closes the channel associated with the given index.
func (at *Await[T]) Close(idx int64) (err error) {
	defer recoverpanic(&err) // Recover from panic if it occurs
	loop := 1000
START:
	if at.m.Has(idx) {
		at.mux.Lock(idx)
		defer at.mux.Unlock(idx)
		if ch, ok := at.m.Get(idx); ok {
			at.m.Del(idx) // Remove the channel from the map
			close(ch)     // Close the channel
		}
	} else if loop > 0 {
		loop--
		<-time.After(time.Millisecond)
		goto START
	}
	return
}

// CloseAndPut sends a value to the channel and deletes it from the map.
func (at *Await[T]) CloseAndPut(idx int64, v T) (err error) {
	defer recoverpanic(&err) // Recover from panic if it occurs
	loop := 1000
START:
	if at.m.Has(idx) {
		at.mux.Lock(idx)
		defer at.mux.Unlock(idx)
		if ch, ok := at.m.Get(idx); ok {
			at.m.Del(idx) // Remove the channel from the map
			ch <- v       // Send value to the channel
			close(ch)
		}
	} else if loop > 0 {
		loop--
		<-time.After(time.Millisecond)
		goto START
	}
	return
}

// Wait wait for the channel data to return or close, and set the timeout period
func (at *Await[T]) Wait(idx int64, timeout time.Duration) (r T, err error) {
	defer recoverpanic(&err)
	defer at.m.Del(idx)
	ch := at.Get(idx)
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		select {
		case <-timer.C:
			defer func() {
				defer recoverpanic(nil)
				close(ch)
			}()
			return r, fmt.Errorf("wait %d timeout", idx)
		case r = <-ch:
			return r, nil
		}
	} else {
		r = <-ch
	}
	return
}
