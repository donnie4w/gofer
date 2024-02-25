//go:build !windows && !wasm
// +build !windows,!wasm

package mmap

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

func mmapSyncToDisk(file *os.File, mappedMemory []byte) (err error) {
	ptr := unsafe.Pointer(&mappedMemory[0])
	_, _, errno := unix.Syscall(unix.SYS_MSYNC, uintptr(ptr), uintptr(len(mappedMemory)), unix.MS_SYNC)
	if errno != 0 {
		err = fmt.Errorf("msync failed: %v", errno)
	}
	return
}