// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/mmap

//go:build solaris
// +build solaris

package mmap

import (
	"os"
)

func mmapSyncToDisk(file *os.File, mappedMemory []byte, n int64, length int) (err error) {
	return
}
