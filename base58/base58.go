// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
// github.com/donnie4w/gofer/base58
// https://en.bitcoin.it/wiki/Base58Check_encoding#Version_bytes

package base58

import "bytes"

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// EncodeForInt64 encode uint64 to []byte
func EncodeForInt64(v uint64) (_r []byte) {
	for v > 0 {
		mod := v % 58
		v /= 58
		_r = append(_r, b58Alphabet[mod])
	}
	return reverseBytes(_r)
}

func reverseBytes(bytes []byte) []byte {
	for i := 0; i < len(bytes)/2; i++ {
		bytes[i], bytes[len(bytes)-1-i] = bytes[len(bytes)-1-i], bytes[i]
	}
	return bytes
}

// DecodeForInt64 decode []byte to uint64
func DecodeForInt64(bs []byte) (_r uint64, ok bool) {
	defer func() { recover() }()
	for _, b := range bs {
		if idx := bytes.IndexByte(b58Alphabet, b); idx >= 0 {
			_r *= 58
			_r += uint64(idx)
		} else {
			return 0, false
		}
	}
	ok = true
	return
}
