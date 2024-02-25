//go:build solaris
// +build solaris

package mmap

import (
	"os"
)

func mmapSyncToDisk(file *os.File, mappedMemory []byte) (err error) {
	return
}
