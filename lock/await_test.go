package lock

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAwait_WaitAndCloseAndPut(t *testing.T) {

	at := NewAwait[int](64)

	go func() {
		time.Sleep(10 * time.Millisecond)
		_ = at.CloseAndPut(1, 100)
	}()

	v, err := at.Wait(1, time.Second)

	if err != nil {
		t.Fatal(err)
	}

	if v != 100 {
		t.Fatalf("expect 100 got %d", v)
	}
}

func TestAwait_CloseAndPutBeforeWait(t *testing.T) {

	at := NewAwait[int](64)

	go func() {
		_ = at.CloseAndPut(2, 200)
	}()

	time.Sleep(10 * time.Millisecond)

	v, err := at.Wait(2, time.Second)

	if err != nil {
		t.Fatal(err)
	}

	if v != 200 {
		t.Fatalf("expect 200 got %d", v)
	}
}

func TestAwait_Timeout(t *testing.T) {

	at := NewAwait[int](64)

	start := time.Now()

	_, err := at.Wait(3, 50*time.Millisecond)

	if err == nil {
		t.Fatal("expect timeout error")
	}

	if time.Since(start) < 40*time.Millisecond {
		t.Fatal("timeout too early")
	}
}

func TestAwait_Cancel(t *testing.T) {

	at := NewAwait[int](64)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, isCancel, err := at.WaitWithCancel(ctx, 4, time.Second)

	if err == nil {
		t.Fatal("expect cancel error")
	}

	if !isCancel {
		t.Fatal("expect cancel flag true")
	}
}

func TestAwait_Close(t *testing.T) {

	at := NewAwait[int](64)

	go func() {
		time.Sleep(20 * time.Millisecond)
		_ = at.Close(5)
	}()

	_, err := at.Wait(5, time.Second)

	if err != nil {
		t.Fatal(err)
	}
}

func TestAwait_Has(t *testing.T) {

	at := NewAwait[int](64)

	if at.Has(6) {
		t.Fatal("should not exist")
	}

	ch := at.Get(6)

	if ch == nil {
		t.Fatal("channel nil")
	}

	if !at.Has(6) {
		t.Fatal("should exist")
	}
}

func TestAwait_SyncWait(t *testing.T) {

	at := NewAwait[int](64)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {

		v, err := at.SyncWait(&wg, 7, time.Second)

		if err != nil {
			t.Fatal(err)
		}

		if v != 700 {
			t.Fatalf("expect 700 got %d", v)
		}
	}()

	wg.Wait()

	_ = at.CloseAndPut(7, 700)
}

func TestAwait_ConcurrentRPC(t *testing.T) {

	at := NewAwait[int](64)

	const n = 10000

	var wg sync.WaitGroup

	for i := 0; i < n; i++ {

		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			go func() {
				time.Sleep(time.Microsecond)
				_ = at.CloseAndPut(int64(i), i)
			}()

			v, err := at.Wait(int64(i), time.Second)

			if err != nil {
				t.Fatal(err)
			}

			if v != i {
				t.Fatalf("expect %d got %d", i, v)
			}

		}(i)
	}

	wg.Wait()
}

func TestAwait_HighConcurrency(t *testing.T) {

	at := NewAwait[int](64)

	const n = 50000

	var success int64

	var wg sync.WaitGroup

	for i := 0; i < n; i++ {

		wg.Add(1)

		go func(i int) {

			defer wg.Done()

			go func() {
				time.Sleep(time.Microsecond)
				_ = at.CloseAndPut(int64(i), i)
			}()

			v, err := at.Wait(int64(i), time.Second)

			if err == nil && v == i {
				atomic.AddInt64(&success, 1)
			}

		}(i)
	}

	wg.Wait()

	if success != n {
		t.Fatalf("success %d expect %d", success, n)
	}
}

func BenchmarkAwait_RoundTrip(b *testing.B) {

	at := NewAwait[int](64)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		idx := int64(i)

		go func() {
			_ = at.CloseAndPut(idx, 1)
		}()

		_, _ = at.Wait(idx, time.Second)
	}
}

func BenchmarkAwait_Parallel(b *testing.B) {

	at := NewAwait[int](64)

	b.RunParallel(func(pb *testing.PB) {

		var id int64

		for pb.Next() {

			idx := atomic.AddInt64(&id, 1)

			go func() {
				_ = at.CloseAndPut(idx, 1)
			}()

			_, _ = at.Wait(idx, time.Second)
		}
	})
}
