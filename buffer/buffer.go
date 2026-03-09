// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/buffer

package buffer

import (
	"fmt"
	"io"
	"unsafe"

	gobuffer "github.com/donnie4w/gofer/pool/buffer"
)

var BufPool = gobuffer.NewPool[Buffer](func() *Buffer {
	b := make([]byte, 0)
	return (*Buffer)(&b)
}, func(b *Buffer) { b.Reset() })

func NewBuffer() *Buffer {
	b := make([]byte, 0)
	return (*Buffer)(&b)
}

func NewBufferWithCapacity(capacity int) *Buffer {
	b := make([]byte, 0, capacity)
	return (*Buffer)(&b)
}

func NewBufferByPool() *Buffer {
	return BufPool.Get()
}

func NewBufferBySlice(bs []byte) *Buffer {
	return (*Buffer)(&bs)
}

type Buffer []byte

func (b *Buffer) Reset() {
	if b != nil {
		*b = (*b)[:0]
	}
}

func (b *Buffer) Write(p []byte) (int, error) {
	if b == nil {
		return 0, fmt.Errorf("Write: buffer is nil")
	}
	*b = append(*b, p...)
	return len(p), nil
}

func (b *Buffer) WriteString(s string) (int, error) {
	if b == nil {
		return 0, fmt.Errorf("WriteString: buffer is nil")
	}
	*b = append(*b, s...)
	return len(s), nil
}

func (b *Buffer) WriteInt32(i int) (int, error) {
	if b == nil {
		return 0, fmt.Errorf("WriteInt32: buffer is nil")
	}
	*b = append(*b,
		byte(i>>24),
		byte(i>>16),
		byte(i>>8),
		byte(i),
	)
	return 4, nil
}

func (b *Buffer) WriteByte(c byte) error {
	if b != nil {
		*b = append(*b, c)
		return nil
	} else {
		return fmt.Errorf("WriteByte: buffer is nil")
	}
}

func (b *Buffer) Bytes() []byte {
	if b != nil {
		return []byte(*b)
	}
	return nil
}

func (b *Buffer) Free() {
	if b != nil {
		BufPool.Put(&b)
	}
}

func (b *Buffer) Len() int {
	if b == nil {
		return 0
	}
	return len(*b)
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	if b == nil || len(*b) == 0 {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	if n = copy(p, *b); n < b.Len() {
		*b = (*b)[n:]
	}
	return n, nil
}

func (b *Buffer) String() string {
	if b == nil || len(*b) == 0 {
		return ""
	}
	// 安全的零拷贝转换
	return unsafe.String(unsafe.SliceData(*b), len(*b))
}
