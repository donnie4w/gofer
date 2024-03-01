//go:build solaris
// +build solaris

package mmap

import (
	"os"
)

func mmapSyncToDisk(file *os.File, mappedMemory []byte, n int64, length int) (err error) {
	return
}
