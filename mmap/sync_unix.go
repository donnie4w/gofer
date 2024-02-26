//go:build !windows && !wasm && !solaris
// +build !windows,!wasm,!solaris

package mmap

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

func mmapSyncToDisk(file *os.File, mappedMemory []byte) (err error) {
	pageSize := unix.Getpagesize()
	ptr := unsafe.Pointer(&mappedMemory[0])
	startOffset := uintptr(ptr)
	length := uintptr(len(mappedMemory))
	alignedStart := startOffset &^ (pageSize - 1)
	alignedLength := (length + (startOffset - alignedStart)) &^ (pageSize - 1)
	_, _, errno := unix.Syscall6(unix.SYS_MSYNC, alignedStart, alignedLength, unix.MS_SYNC, 0, 0, 0)
	if errno != 0 {
		return fmt.Errorf("msync failed: %v", errno)
	}
	return nil
}
