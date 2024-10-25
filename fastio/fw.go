// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/fastio

package fastio

import (
	"bufio"
	"sync"
	"sync/atomic"
	"time"
)

type writer interface {
	WriteSync(bs []byte) (offset int64, err error)
	Write(bs []byte) (n int, err error)
}

type writeData struct {
	data   []byte
	offset int64
	done   chan struct{}
}

type mwriter struct {
	ch    chan *writeData
	mu    sync.RWMutex
	wc    int32
	cc    int32
	fh    *fileHandle
	ioend bool
}

func newWriter(fh *fileHandle) writer {
	return &mwriter{fh: fh, ch: make(chan *writeData, 1<<16), ioend: true}
}

func (m *mwriter) WriteSync(bs []byte) (offset int64, err error) {
	defer atomic.AddInt32(&m.cc, -1)
	if atomic.AddInt32(&m.cc, 1) <= 1 {
		m.mu.Lock()
		defer m.mu.Unlock()
		if m.wc > 0 {
			m.ioWrite()
		}
		n := 0
		n, err = m.fh.file.Write(bs)
		offset = m.fh.offset
		atomic.AddInt64(&m.fh.offset, int64(n))
		return
	}
	atomic.AddInt32(&m.wc, 1)
	wd := &writeData{data: bs, done: make(chan struct{})}
	m.ch <- wd
START:
	if m.ioend && m.mu.TryLock() {
		if err = m.ioWrite(); err != nil {
			return
		}
		m.mu.Unlock()
	} else {
		select {
		case <-wd.done:
			goto END
		case <-time.After(time.Microsecond):
			goto START
		}
	}
END:
	offset = wd.offset
	return
}

func (m *mwriter) Write(bs []byte) (n int, err error) {
	defer atomic.AddInt32(&m.cc, -1)
	if atomic.AddInt32(&m.cc, 1) <= 1 {
		m.mu.RLock()
		defer m.mu.RUnlock()
		if m.wc > 0 {
			m.ioWrite()
		}
		if n, err = m.fh.file.Write(bs); n > 0 {
			atomic.AddInt64(&m.fh.offset, int64(n))
		}
		return n, err
	}
	atomic.AddInt32(&m.wc, 1)
	m.ch <- &writeData{data: bs}
START:
	if m.ioend && m.mu.TryLock() {
		if err = m.ioWrite(); err != nil {
			return
		}
		m.mu.Unlock()
	} else if m.ioend {
		goto START
	}
	return len(bs), nil
}

func (m *mwriter) ioWrite() (err error) {
	if m.wc == 0 {
		return
	}
	m.ioend = false
	defer recoverable(&err)
	bw := bufio.NewWriter(m.fh.file)
	for wd := range m.ch {
		if wd.done != nil {
			defer func() { close(wd.done) }()
		}
		wd.offset = m.fh.offset
		if _, err = bw.Write(wd.data); err != nil {
			return
		}
		atomic.AddInt64(&m.fh.offset, int64(len(wd.data)))
		if atomic.AddInt32(&m.wc, -1) == 0 {
			break
		}
	}
	m.ioend = true
	err = bw.Flush()
	return
}
