// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/util

package util

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"github.com/golang/snappy"
	"io"
)

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
