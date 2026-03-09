// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/util

package util

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/donnie4w/gofer/uuid"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/exp/mmap"
	"hash/crc32"
	"hash/crc64"
	"hash/fnv"
	"hash/maphash"
	mathrand "math/rand"
	"os"
	"regexp"
	"strings"
)

func Concat(ss ...string) string {
	totalLen := 0
	for _, s := range ss {
		totalLen += len(s)
	}
	builder := strings.Builder{}
	builder.Grow(totalLen)
	for _, v := range ss {
		builder.WriteString(v)
	}
	return builder.String()
}

func MD5(bs []byte) []byte {
	if len(bs) == 0 {
		return nil
	}
	m := md5.New()
	_, _ = m.Write(bs)
	return m.Sum(nil)
}

func Md5Str(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(MD5([]byte(s))))
}

func SHA1(bs []byte) []byte {
	if len(bs) == 0 {
		return nil
	}
	m := sha1.New()
	_, _ = m.Write(bs)
	return m.Sum(nil)
}

//func Sha1Str(s string) string {
//	if s == "" {
//		return ""
//	}
//	return strings.ToUpper(hex.EncodeToString(SHA1([]byte(s))))
//}

var crc8Tab = []byte{0, 7, 14, 9, 28, 27, 18, 21, 56, 63, 54, 49, 36, 35, 42, 45, 112, 119, 126, 121, 108, 107, 98, 101, 72, 79, 70, 65, 84, 83, 90, 93, 224, 231, 238, 233, 252, 251, 242, 245, 216, 223, 214, 209, 196, 195, 202, 205, 144, 151, 158, 153, 140, 139, 130, 133, 168, 175, 166, 161, 180, 179, 186, 189, 199, 192, 201, 206, 219, 220, 213, 210, 255, 248, 241, 246, 227, 228, 237, 234, 183, 176, 185, 190, 171, 172, 165, 162, 143, 136, 129, 134, 147, 148, 157, 154, 39, 32, 41, 46, 59, 60, 53, 50, 31, 24, 17, 22, 3, 4, 13, 10, 87, 80, 89, 94, 75, 76, 69, 66, 111, 104, 97, 102, 115, 116, 125, 122, 137, 142, 135, 128, 149, 146, 155, 156, 177, 182, 191, 184, 173, 170, 163, 164, 249, 254, 247, 240, 229, 226, 235, 236, 193, 198, 207, 200, 221, 218, 211, 212, 105, 110, 103, 96, 117, 114, 123, 124, 81, 86, 95, 88, 77, 74, 67, 68, 25, 30, 23, 16, 5, 2, 11, 12, 33, 38, 47, 40, 61, 58, 51, 52, 78, 73, 64, 71, 82, 85, 92, 91, 118, 113, 120, 127, 106, 109, 100, 99, 62, 57, 48, 55, 34, 37, 44, 43, 6, 1, 8, 15, 26, 29, 20, 19, 174, 169, 160, 167, 178, 181, 188, 187, 150, 145, 152, 159, 138, 141, 132, 131, 222, 217, 208, 215, 194, 197, 204, 203, 230, 225, 232, 239, 250, 253, 244, 243}

func CRC8(bs []byte) (_r byte) {
	if len(bs) == 0 {
		return
	}
	for _, v := range bs {
		_r = crc8Tab[_r^v]
	}
	return
}

var crc32Table = crc32.MakeTable(crc32.IEEE)

func CRC32(bs []byte) uint32 {
	if len(bs) == 0 {
		return 0
	}
	return crc32.Checksum(bs, crc32Table)
}

var crc64ECMATable = crc64.MakeTable(crc64.ECMA)

func CRC64(bs []byte) uint64 {
	return crc64.Checksum(bs, crc64ECMATable)
}

var seed = maphash.MakeSeed()

func Hash64(key []byte) uint64 {
	return maphash.Bytes(seed, key)
}

func FNVHash32(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(data)
	return h.Sum32()
}

func FNVHash64(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}

var mathrng = mathrand.New(mathrand.NewSource(uuid.NewUUID().Int64()))

// RandUint 伪随机
func RandUint(i uint) uint {
	return uint(mathrng.Intn(int(i)))
}

// RandUintCrypto 真随机
func RandUintCrypto(i uint) uint {
	if i == 0 {
		return 0
	}
	var n uint
	err := binary.Read(rand.Reader, binary.BigEndian, &n)
	if err != nil {
		return RandUint(i)
	}
	return n
}

func MatchString(pattern string, s string) bool {
	b, err := regexp.MatchString(pattern, s)
	if err != nil {
		b = false
	}
	return b
}

//func TimeParse(s string) (t time.Time, err error) {
//	if s == "" {
//		return time.Time{}, fmt.Errorf("time string is empty")
//	}
//	loc, err := time.LoadLocation("Asia/Shanghai")
//	if err != nil {
//		return time.Time{}, fmt.Errorf("load location failed: %w", err)
//	}
//	t, err = time.ParseInLocation("2006-01-02 15:04:05", s, loc)
//	if err != nil {
//		return time.Time{}, fmt.Errorf("parse time failed: %w", err)
//	}
//	return t, nil
//	return
//}

func UUID64() int64 {
	return uuid.NewUUID().Int64()
}

func UUID32() uint32 {
	return uuid.NewUUID().Uint32()
}

func IsFileExist(path string) (_r bool) {
	if path != "" {
		_, err := os.Stat(path)
		_r = err == nil || os.IsExist(err)
	}
	return
}

func ReadFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("file path is empty")
	}
	f, err := mmap.Open(path)
	if err != nil {
		return nil, fmt.Errorf("mmap open failed: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()
	fileLen := f.Len()
	if fileLen == 0 {
		return []byte{}, nil
	}
	bs := make([]byte, fileLen)
	n, err := f.ReadAt(bs, 0)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}
	if n != fileLen {
		return nil, fmt.Errorf("read incomplete: read %d bytes, expected %d", n, fileLen)
	}
	return bs, nil
}

func Recover(err *error) {
	if e := recover(); e != nil {
		if err != nil {
			*err = fmt.Errorf("%v", e)
		}
	}
}

// Blake2b 计算字节数组的Blake2b-256哈希值
func Blake2b(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("input is empty")
	}
	hasher, err := blake2b.New256(nil)
	if err != nil {
		return nil, fmt.Errorf("create blake2b hasher failed: %w", err)
	}
	_, _ = hasher.Write(input)
	return hasher.Sum(nil), nil
}
