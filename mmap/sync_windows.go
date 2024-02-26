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
