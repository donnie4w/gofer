package buffer

import (
	gobuffer "github.com/donnie4w/gofer/pool/buffer"
)

var BufPool = gobuffer.NewPool[Buffer](func() *Buffer {
	b := make([]byte, 0)
	return (*Buffer)(&b)
}, func(b *Buffer) { b.Reset() })

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

func NewBuffer() *Buffer {
	b := make([]byte, 0)
	return (*Buffer)(&b)
}

func NewBufferByPool() *Buffer {
	return BufPool.Get()
}

func (b *Buffer) Free() {
	if cap(*b) <= 1<<14 {
		BufPool.Put(&b)
	}
}

func (b *Buffer) Len() int {
	return len([]byte(*b))
}

func int32ToBytes(n int32) (bs []byte) {
	bs = make([]byte, 4)
	for i := 0; i < 4; i++ {
		bs[i] = byte(n >> (8 * (3 - i)))
	}
	return
}
