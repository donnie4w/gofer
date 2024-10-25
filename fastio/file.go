// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/fastio

package fastio

import (
	"os"
)

type File interface {
	Write([]byte) (int, error)
	WriteAt([]byte, int64) (int, error)
	WriteSync([]byte) (int64, error)
	ReadAt([]byte, int64) (int, error)
	Read([]byte) (int, error)
	Close() error
}

func Open(path string) (r File, err error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return New(f)
}

func New(f *os.File) (r File, err error) {
	ret, err := offset(f)
	if err != nil {
		return nil, err
	}
	fh := &fileHandle{file: f, offset: ret}
	fh.writer = newWriter(fh)
	return fh, nil
}

type fileHandle struct {
	offset  int64
	file    *os.File
	isClose bool
	writer  writer
}

func (f *fileHandle) WriteAt(b []byte, off int64) (n int, err error) {
	return f.file.WriteAt(b, off)
}

func (f *fileHandle) WriteSync(b []byte) (offset int64, err error) {
	return f.writer.WriteSync(b)
}

func (f *fileHandle) Write(b []byte) (n int, err error) {
	return f.writer.Write(b)
}

func (f *fileHandle) ReadAt(b []byte, off int64) (n int, err error) {
	return f.file.ReadAt(b, off)
}

func (f *fileHandle) Read(b []byte) (n int, err error) {
	return f.file.Read(b)
}

func (f *fileHandle) Close() (err error) {
	f.isClose = true
	return f.file.Close()
}
