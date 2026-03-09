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
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(n))
	return b
}

func BytesToInt64(bs []byte) (_r int64) {
	if len(bs) == 0 {
		return
	}

	b := make([]byte, 8)
	copy(b[8-len(bs):], bs[:min(len(bs), 8)])
	return int64(binary.BigEndian.Uint64(b))
}

func Int32ToBytes(n int32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(n))
	return b
}

func BytesToInt32(bs []byte) (_r int32) {
	if len(bs) == 0 {
		return
	}

	b := make([]byte, 4)
	copy(b[4-len(bs):], bs[:min(len(bs), 4)])
	return int32(binary.BigEndian.Uint32(b))
}

func Int16ToBytes(n int16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(n))
	return b
}

func BytesToInt16(bs []byte) (_r int16) {
	if len(bs) == 0 {
		return 0
	}

	b := make([]byte, 2)
	copy(b[2-len(bs):], bs[:min(len(bs), 2)])
	return int16(binary.BigEndian.Uint16(b))
}

func IntArrayToBytes(n []int64) []byte {
	if len(n) == 0 {
		return nil
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(n)*8)) // 预分配容量，减少扩容
	if err := binary.Write(buf, binary.BigEndian, n); err != nil {
		return nil
	}
	return buf.Bytes()
}

func BytesToIntArray(bs []byte) (data []int64) {
	if len(bs) == 0 {
		return nil
	}
	if len(bs)%8 != 0 {
		return nil
	}

	data = make([]int64, len(bs)/8)
	buf := bytes.NewBuffer(bs)
	if err := binary.Read(buf, binary.BigEndian, &data); err != nil {
		return nil
	}
	return
}
