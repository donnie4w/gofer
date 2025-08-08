// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compress

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"

	"github.com/klauspost/compress/snappy"
	"github.com/klauspost/compress/zstd"
)

func Zlib(input []byte) ([]byte, error) {
	var b bytes.Buffer
	zw := zlib.NewWriter(&b)
	if _, err := zw.Write(input); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func ZlibLevel(bs []byte, level int) (_r []byte, err error) {
	var buf bytes.Buffer
	var compressor *zlib.Writer
	if compressor, err = zlib.NewWriterLevel(&buf, level); err == nil {
		defer compressor.Close()
		compressor.Write(bs)
		compressor.Flush()
		_r = buf.Bytes()
	} else {
		_r = bs
	}
	return
}

func UnZlib(compressed []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var decompressedData bytes.Buffer
	if _, err := decompressedData.ReadFrom(r); err != nil {
		return nil, err
	}

	return decompressedData.Bytes(), nil
}

func Gzip(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	if _, err := gzWriter.Write(input); err != nil {
		return nil, err
	}
	if err := gzWriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func UnGzip(compressed []byte) ([]byte, error) {
	gzReader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	var decompressedData bytes.Buffer
	if _, err := io.Copy(&decompressedData, gzReader); err != nil {
		return nil, err
	}

	return decompressedData.Bytes(), nil
}

func Snappy(bs []byte) (_r []byte) {
	return snappy.Encode(nil, bs)
}

func UnSnappy(bs []byte) (_r []byte, err error) {
	_r, err = snappy.Decode(nil, bs)
	return
}

// Zstd 压缩（使用默认级别）
func Zstd(bs []byte) (_r []byte, err error) {
	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		return nil, err
	}
	defer encoder.Close()

	// 使用 EncodeAll，预分配适当容量
	dst := make([]byte, 0, len(bs)) // 可进一步优化为 len(bs)*0.5 ~ 0.8
	return encoder.EncodeAll(bs, dst), nil
}

// ZstdLevel 压缩（指定压缩级别）
func ZstdLevel(bs []byte, level zstd.EncoderLevel) (_r []byte, err error) {
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(level))
	if err != nil {
		return nil, err
	}
	defer encoder.Close()

	dst := make([]byte, 0, len(bs))
	return encoder.EncodeAll(bs, dst), nil
}

// UnZstd 解压
func UnZstd(bs []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
	if err != nil {
		return nil, err
	}
	defer decoder.Close()

	return decoder.DecodeAll(bs, nil)
}
