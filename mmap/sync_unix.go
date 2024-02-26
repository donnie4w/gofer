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
	ptr := unsafe.Pointer(&mappedMemory[0])
	length := uintptr(len(mappedMemory))

	_, _, errno := unix.Syscall6(unix.SYS_MSYNC, uintptr(ptr), length, unix.MS_SYNC, 0, 0, 0)
	if errno != 0 {
		return fmt.Errorf("msync failed: %v", errno)
	}

	return nil
}
