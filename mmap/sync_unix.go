//go:build !windows && !wasm && !solaris
// +build !windows,!wasm,!solaris

package mmap

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

func mmapSyncToDisk(file *os.File, mappedMemory []byte, n int64, length int) (err error) {
	pageSize := int64(unix.Getpagesize())
	var ptr unsafe.Pointer
	var leng uintptr
	p := n / pageSize
	alignedStart := p * pageSize
	ptr = unsafe.Pointer(&mappedMemory[int(alignedStart)])
	leng = uintptr(int(n-alignedStart) + length)
	_, _, errno := unix.Syscall6(unix.SYS_MSYNC, uintptr(ptr), leng, unix.MS_SYNC, 0, 0, 0)
	if errno != 0 {
		return fmt.Errorf("msync failed: %v", errno)
	}
	return nil
}
