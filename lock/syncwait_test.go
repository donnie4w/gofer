// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/lock

package lock

import (
	"fmt"
	"github.com/donnie4w/gofer/uuid"
	"math/rand"
	"testing"
	"time"
)

func TestSyncWait(t *testing.T) {
	// Initialize SyncWait instance with a limit for locks
	aw := NewSyncWait(10)

	// Test basic waiting functionality
	idx := int64(1)
	go func() {
		// Simulate some work that takes 1 second
		time.Sleep(1 * time.Second)
		// Mark the task as done
		aw.Close(idx)
	}()

	// Call Wait and expect it to complete without any errors
	aw.Wait(idx) // This should not time out, as the task completes within 1 second
	t.Log("Task completed successfully")

	// Test WaitWithTimeOut with a timeout
	go func() {
		// Simulate some work that takes 2 seconds
		time.Sleep(2 * time.Second)
		// Mark the task as done
		aw.Close(idx)
	}()

	// Set timeout to 1 second, expecting a timeout error
	err := aw.WaitWithTimeOut(idx, 1*time.Second)
	if err == nil || err.Error() != fmt.Sprintf("timeout after %s", 1*time.Second) {
		t.Errorf("Expected timeout error, but got: %v", err)
	}
	t.Log("Timeout test passed")
}

func BenchmarkSyncWait(b *testing.B) {
	// Initialize SyncWait instance with a limit for locks
	aw := NewSyncWait(100)

	// Benchmark with 100 concurrent tasks
	b.Run("Benchmark 100 tasks", func(b *testing.B) {
		// Run the benchmark for 100 times
		b.ReportAllocs() // Report memory allocations
		for i := 0; i < b.N; i++ {
			idx := int64(i)
			// Launch 100 concurrent tasks
			go func(idx int64) {
				// Simulate some work that takes random time between 100ms to 500ms
				time.Sleep(time.Duration(10+rand.Intn(40)) * time.Nanosecond)
				// Mark task as done
				aw.Close(idx)
			}(idx)

			// Wait for the task to finish
			aw.Wait(idx)
		}
	})

	// Benchmark with 100 tasks and timeout
	b.Run("Benchmark 100 tasks with timeout", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			idx := uuid.NewUUID().Int64()
			go func(idx int64) {
				// Simulate work
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Nanosecond)
				aw.Close(idx)
			}(idx)

			err := aw.WaitWithTimeOut(idx, time.Second)
			if err != nil {
				b.Errorf("Error waiting with timeout: %v", err)
			}
		}
	})
}
