//go:build !windows && !wasm
// +build !windows,!wasm

package mmap

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func mmapSyncToDisk(file *os.File, mappedMemory []byte) (err error) {
	ptr := unsafe.Pointer(&mappedMemory[0])
	_, _, errno := syscall.SyscallN(syscall.SYS_MSYNC, uintptr(ptr), uintptr(len(mappedMemory)), syscall.MS_SYNC)
	if errno != 0 {
		err = fmt.Errorf("msync failed: %v", errno)
	}
	return
}
