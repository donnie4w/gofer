// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/mmap

package mmap

import (
	"errors"
	"os"
	"sync"

	gommap "github.com/edsrzf/mmap-go"
)

func NewMMAP(f *os.File, startOffset int64) (m *Mmap, err error) {
	fif, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if fif.Size() == 0 {
		return nil, errors.New("the file capacity is zero")
	}

	_mmap, err := gommap.Map(f, gommap.RDWR, 0)
	if err != nil {
		return nil, err
	}
	if startOffset >= fif.Size() {
		return nil, errors.New("offset Exceeds the file limit")
	}
	m = &Mmap{file: f, _mmap: _mmap, maxsize: fif.Size(), mux: &sync.Mutex{}, offset: startOffset}
	return
}

type Mmap struct {
	file    *os.File
	_mmap   gommap.MMap
	offset  int64
	maxsize int64
	mux     *sync.Mutex
}

func (t *Mmap) Append(bs []byte) (n int64, err error) {
	t.mux.Lock()
	if t.offset+int64(len(bs)) > t.maxsize {
		err = errors.New("exceeding file size limit")
		t.mux.Unlock()
		return
	}
	n = t.offset
	t.offset += int64(len(bs))
	t.mux.Unlock()
	copy(t._mmap[n:int(n)+len(bs)], bs)
	return
}

func (t *Mmap) AppendSync(bs []byte) (n int64, err error) {
	if n, err = t.Append(bs); err == nil {
		err = mmapSyncToDisk(t.file, t._mmap, n, len(bs))
	}
	return
}

func (t *Mmap) WriteAt(bs []byte, offset int) (err error) {
	t.mux.Lock()
	if offset+len(bs) > int(t.maxsize) {
		err = errors.New("exceeding file size limit")
		t.mux.Unlock()
		return
	}
	if int64(offset+len(bs)) > t.offset {
		t.offset = int64(offset + len(bs))
	}
	t.mux.Unlock()
	copy(t._mmap[offset:offset+len(bs)], bs)
	return
}

func (t *Mmap) WriteAtSync(bs []byte, offset int) (err error) {
	if err = t.WriteAt(bs, offset); err == nil {
		err = mmapSyncToDisk(t.file, t._mmap, int64(offset), len(bs))
	}
	return
}

func (t *Mmap) Unmap() error {
	return t._mmap.Unmap()
}

func (t *Mmap) UnmapAndCloseFile() (err error) {
	if err = t._mmap.Unmap(); err == nil {
		err = t.file.Close()
	}
	return
}

func (t *Mmap) Flush() error {
	return t._mmap.Flush()
}

func (t *Mmap) Bytes() []byte {
	return t._mmap
}

func (t *Mmap) FileSize() int64 {
	return t.maxsize
}

func (t *Mmap) Close() (err error) {
	if err = t.Flush(); err == nil {
		err = t.Unmap()
	}
	return
}
