//go:build windows || wasm
// +build windows wasm

package mmap

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

func mmapSyncToDisk(file *os.File, mappedMemory []byte) (err error) {
	hFile := windows.Handle(file.Fd())
	err = windows.FlushViewOfFile(uintptr(unsafe.Pointer(&mappedMemory[0])), uintptr(len(mappedMemory)))
	if err == nil {
		err = windows.FlushFileBuffers(hFile)
	}
	return
}
