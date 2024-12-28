// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/util

package util

import (
	"bytes"
	"encoding/binary"
)

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
