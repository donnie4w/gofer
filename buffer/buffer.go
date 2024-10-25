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
	*b = (*b)[:0]
}

func (b *Buffer) Write(p []byte) (int, error) {
	if b != nil {
		*b = append(*b, p...)
		return len(p), nil
	} else {
		return 0, fmt.Errorf("Write: buffer is nil")
	}
}

func (b *Buffer) WriteString(s string) (int, error) {
	if b != nil {
		*b = append(*b, s...)
		return len(s), nil
	} else {
		return 0, fmt.Errorf("WriteString: buffer is nil")
	}
}

func (b *Buffer) WriteInt32(i int) (int, error) {
	if b != nil {
		*b = append(*b, int32ToBytes(int32(i))...)
		return 8, nil
	} else {
		return 0, fmt.Errorf("WriteInt32: buffer is nil")
	}
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
	return len([]byte(*b))
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	if b == nil {
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	if n = copy(p, *b); n < b.Len() {
		*b = (*b)[n:]
	}
	return n, nil
}

func (b *Buffer) String() string {
	if b != nil {
		return string(b.Bytes())
	}
	return ""
}

func int32ToBytes(n int32) (bs []byte) {
	bs = make([]byte, 4)
	for i := 0; i < 4; i++ {
		bs[i] = byte(n >> (8 * (3 - i)))
	}
	return
}
