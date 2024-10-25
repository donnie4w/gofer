// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/mmap

//go:build windows || wasm
// +build windows wasm

package mmap

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

func mmapSyncToDisk(file *os.File, mappedMemory []byte, n int64, length int) (err error) {
	hFile := windows.Handle(file.Fd())
	bs := mappedMemory[n:(n + int64(length))]
	err = windows.FlushViewOfFile(uintptr(unsafe.Pointer(&bs[0])), uintptr(len(bs)))
	if err == nil {
		err = windows.FlushFileBuffers(hFile)
	}
	return
}
