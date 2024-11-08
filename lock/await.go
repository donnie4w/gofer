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

// DelAndClose deletes and closes the channel associated with the given index.
func (at *Await[T]) DelAndClose(idx int64) {
	defer recoverpanic() // Recover from panic if it occurs
	if at.m.Has(idx) {
		at.mux.Lock(idx)
		defer at.mux.Unlock(idx)
		if ch, ok := at.m.Get(idx); ok {
			at.m.Del(idx) // Remove the channel from the map
			close(ch)     // Close the channel
		}
	}
}

// DelAndPut sends a value to the channel and deletes it from the map.
func (at *Await[T]) DelAndPut(idx int64, v T) {
	defer recoverpanic() // Recover from panic if it occurs
	if at.m.Has(idx) {
		at.mux.Lock(idx)
		defer at.mux.Unlock(idx)
		if ch, ok := at.m.Get(idx); ok {
			ch <- v       // Send value to the channel
			at.m.Del(idx) // Remove the channel from the map
			close(ch)
		}
	}
}

// Wait wait for the channel data to return or close, and set the timeout period
func (at *Await[T]) Wait(idx int64, t time.Duration) (r T, err error) {
	defer recoverpanic()
	ch := at.Get(idx)
	select {
	case <-time.After(t):
		close(ch)
		return r, fmt.Errorf("timeout")
	case r = <-ch:
		return r, nil
	}
	return
}
