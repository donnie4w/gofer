package mmap

import (
	"errors"
	"os"
	"sync"

	gommap "github.com/edsrzf/mmap-go"
)

func NewMMAP(f *os.File, startOffset int64) (m *Mmap, err error) {
	fif, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if fif.Size() == 0 {
		return nil, errors.New("The file capacity is zero")
	}

	_mmap, err := gommap.Map(f, gommap.RDWR, 0)
	if err != nil {
		return nil, err
	}
	if startOffset >= fif.Size() {
		return nil, errors.New("Offset Exceeds the file limit")
	}
	m = &Mmap{File: f, _mmap: _mmap, maxsize: fif.Size(), mux: sync.Mutex{}, offset: startOffset}
	return
}

type Mmap struct {
	File    *os.File
	_mmap   gommap.MMap
	offset  int64
	maxsize int64
	mux     sync.Mutex
}

func (this *Mmap) Append(bs []byte) (n int64, err error) {
	this.mux.Lock()
	if this.offset+int64(len(bs)) > int64(int(this.maxsize)) {
		err = errors.New("Exceeding file size limit")
		return
	}
	n = this.offset
	this.offset += int64(len(bs))
	this.mux.Unlock()
	copy(this._mmap[n:int(n)+len(bs)], bs)
	return
}

func (this *Mmap) AppendSync(bs []byte) (n int64, err error) {
	if n, err = this.Append(bs); err == nil {
		err = mmapSyncToDisk(this.File, this._mmap[n:int(n)+len(bs)])
	}
	return
}

func (this *Mmap) Write(bs []byte, offset int) (err error) {
	this.mux.Lock()
	if offset+len(bs) > int(this.maxsize) {
		err = errors.New("Exceeding file size limit")
		return
	}
	if int64(offset+len(bs)) > this.offset {
		this.offset = int64(offset + len(bs))
	}
	this.mux.Unlock()
	copy(this._mmap[offset:offset+len(bs)], bs)
	return
}

func (this *Mmap) WriteSync(bs []byte, offset int) (err error) {
	if err = this.Write(bs, offset); err == nil {
		err = mmapSyncToDisk(this.File, this._mmap[offset:offset+len(bs)])
	}
	return
}

func (this *Mmap) Unmap() error {
	return this._mmap.Unmap()
}

func (this *Mmap) Flush() error {
	return this._mmap.Flush()
}

func (this *Mmap) Bytes() []byte {
	return this._mmap
}

func (this *Mmap) Close() (err error) {
	if err = this.Flush(); err == nil {
		err = this.Unmap()
	}
	return
}
