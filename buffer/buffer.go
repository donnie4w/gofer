// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/gofer/buffer

package buffer

import (
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
	*b = append(*b, p...)
	return len(p), nil
}

func (b *Buffer) WriteString(s string) (int, error) {
	*b = append(*b, s...)
	return len(s), nil
}

func (b *Buffer) WriteInt32(i int) (int, error) {
	*b = append(*b, int32ToBytes(int32(i))...)
	return 8, nil
}

func (b *Buffer) WriteByte(c byte) error {
	*b = append(*b, c)
	return nil
}

func (b *Buffer) Bytes() []byte {
	return []byte(*b)
}

func (b *Buffer) Free() {
	BufPool.Put(&b)
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
