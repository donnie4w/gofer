// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
// github.com/donnie4w/gofer/util

package util

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"io"

	"github.com/donnie4w/gothrift/thrift"

	"github.com/golang/snappy"
)

func JsonEncode(v any) (bs []byte) {
	bs, _ = json.Marshal(v)
	return
}

func JsonDecode[T any](bs []byte) (_r T, err error) {
	err = json.Unmarshal(bs, &_r)
	return
}

func Encode(e any) (by []byte, err error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(e)
	by = buf.Bytes()
	return
}

func Decode[T any](buf []byte) (_r *T, err error) {
	decoder := gob.NewDecoder(bytes.NewReader(buf))
	_r = new(T)
	err = decoder.Decode(_r)
	return
}

func Int64ToBytes(n int64) []byte {
	var bs [8]byte
	for i := 0; i < 8; i++ {
		bs[i] = byte(n >> (8 * (7 - i)))
	}
	return bs[:]
}

func BytesToInt64(bs []byte) (_r int64) {
	if len(bs) >= 8 {
		for i := 0; i < 8; i++ {
			_r = _r | int64(bs[i])<<(8*(7-i))
		}
	} else {
		var bs8 [8]byte
		for i, b := range bs {
			bs8[i+8-len(bs)] = b
		}
		_r = BytesToInt64(bs8[:])
	}
	return
}

func Int32ToBytes(n int32) []byte {
	var bs [4]byte
	for i := 0; i < 4; i++ {
		bs[i] = byte(n >> (8 * (3 - i)))
	}
	return bs[:]
}

func Int16ToBytes(n int16) []byte {
	var bs [2]byte
	for i := 0; i < 2; i++ {
		bs[i] = byte(n >> (8 * (1 - i)))
	}
	return bs[:]
}

func BytesToInt32(bs []byte) (_r int32) {
	if len(bs) >= 4 {
		for i := 0; i < 4; i++ {
			_r = _r | int32(bs[i])<<(8*(3-i))
		}
	} else {
		bs4 := make([]byte, 4)
		for i, b := range bs {
			bs4[i+4-len(bs)] = b
		}
		_r = BytesToInt32(bs4)
	}
	return
}

func BytesToInt16(bs []byte) (_r int16) {
	if len(bs) >= 2 {
		for i := 0; i < 2; i++ {
			_r = _r | int16(bs[i])<<(8*(1-i))
		}
	} else {
		var bs2 [2]byte
		for i, b := range bs {
			bs2[i+2-len(bs)] = b
		}
		_r = BytesToInt16(bs2[:])
	}
	return
}

func IntArrayToBytes(n []int64) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes()
}

func BytesToIntArray(bs []byte) (data []int64) {
	bytesBuffer := bytes.NewBuffer(bs)
	data = make([]int64, len(bs)/8)
	binary.Read(bytesBuffer, binary.BigEndian, data)
	return
}

var tconf = &thrift.TConfiguration{}

func TEncode(ts thrift.TStruct) (_r []byte) {
	buf := thrift.NewTMemoryBuffer()
	protocol := thrift.NewTCompactProtocolConf(buf, tconf)
	ts.Write(context.TODO(), protocol)
	protocol.Flush(context.TODO())
	_r = buf.Bytes()
	return
}

func TDecode[T thrift.TStruct](bs []byte, ts T) (_r T, err error) {
	buf := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(bs)}
	protocol := thrift.NewTCompactProtocolConf(buf, tconf)
	err = ts.Read(context.TODO(), protocol)
	return ts, err
}

func Zlib(bs []byte) (_r []byte, err error) {
	var buf bytes.Buffer
	var compressor *zlib.Writer
	if compressor, err = zlib.NewWriterLevel(&buf, zlib.BestCompression); err == nil {
		defer compressor.Close()
		compressor.Write(bs)
		compressor.Flush()
		_r = buf.Bytes()
	} else {
		_r = bs
	}
	return
}

func UnZlib(bs []byte) (_r []byte, err error) {
	var obuf bytes.Buffer
	var read io.ReadCloser
	if read, err = zlib.NewReader(bytes.NewReader(bs)); err == nil {
		defer read.Close()
		io.Copy(&obuf, read)
		_r = obuf.Bytes()
	}
	return
}

func Gzip(bs []byte) (buf bytes.Buffer, err error) {
	gw := gzip.NewWriter(&buf)
	defer gw.Close()
	_, err = gw.Write(bs)
	return
}

func UnGzip(bs []byte) (_bb bytes.Buffer, err error) {
	if gz, er := gzip.NewReader(bytes.NewBuffer(bs)); er == nil {
		defer gz.Close()
		var bs = make([]byte, 1024)
		for {
			var n int
			n, err := gz.Read(bs)
			if (err != nil && err != io.EOF) || n == 0 {
				break
			}
			_bb.Write(bs[:n])
		}
	} else {
		err = er
	}
	return
}

func CheckGzipType(bs []byte) (err error) {
	if buf := bytes.NewReader(bs); buf != nil {
		_, err = gzip.NewReader(buf)
	}
	return
}

func SnappyEncode(bs []byte) (_r []byte) {
	return snappy.Encode(nil, bs)
}

func SnappyDecode(bs []byte) (_r []byte) {
	_r, _ = snappy.Decode(nil, bs)
	return
}
