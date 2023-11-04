// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
// github.com/donnie4w/gofer/util

package util

import (
	"crypto/md5"
	crand "crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"hash/crc32"
	"hash/crc64"
	"hash/maphash"
	"math/big"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/exp/mmap"
)

func Concat(ss ...string) string {
	var builder strings.Builder
	for _, v := range ss {
		builder.WriteString(v)
	}
	return builder.String()
}

func MD5(bs []byte) []byte {
	m := md5.New()
	m.Write(bs)
	return m.Sum(nil)
}

func Md5Str(s string) string {
	return strings.ToUpper(hex.EncodeToString(MD5([]byte(s))))
}

func SHA1(bs []byte) []byte {
	m := sha1.New()
	m.Write(bs)
	return m.Sum(nil)
}

func Sha1Str(s string) string {
	return strings.ToUpper(hex.EncodeToString(SHA1([]byte(s))))
}

var crc8Tab = []byte{0, 7, 14, 9, 28, 27, 18, 21, 56, 63, 54, 49, 36, 35, 42, 45, 112, 119, 126, 121, 108, 107, 98, 101, 72, 79, 70, 65, 84, 83, 90, 93, 224, 231, 238, 233, 252, 251, 242, 245, 216, 223, 214, 209, 196, 195, 202, 205, 144, 151, 158, 153, 140, 139, 130, 133, 168, 175, 166, 161, 180, 179, 186, 189, 199, 192, 201, 206, 219, 220, 213, 210, 255, 248, 241, 246, 227, 228, 237, 234, 183, 176, 185, 190, 171, 172, 165, 162, 143, 136, 129, 134, 147, 148, 157, 154, 39, 32, 41, 46, 59, 60, 53, 50, 31, 24, 17, 22, 3, 4, 13, 10, 87, 80, 89, 94, 75, 76, 69, 66, 111, 104, 97, 102, 115, 116, 125, 122, 137, 142, 135, 128, 149, 146, 155, 156, 177, 182, 191, 184, 173, 170, 163, 164, 249, 254, 247, 240, 229, 226, 235, 236, 193, 198, 207, 200, 221, 218, 211, 212, 105, 110, 103, 96, 117, 114, 123, 124, 81, 86, 95, 88, 77, 74, 67, 68, 25, 30, 23, 16, 5, 2, 11, 12, 33, 38, 47, 40, 61, 58, 51, 52, 78, 73, 64, 71, 82, 85, 92, 91, 118, 113, 120, 127, 106, 109, 100, 99, 62, 57, 48, 55, 34, 37, 44, 43, 6, 1, 8, 15, 26, 29, 20, 19, 174, 169, 160, 167, 178, 181, 188, 187, 150, 145, 152, 159, 138, 141, 132, 131, 222, 217, 208, 215, 194, 197, 204, 203, 230, 225, 232, 239, 250, 253, 244, 243}

func CRC8(bs []byte) (_r byte) {
	for _, v := range bs {
		_r = crc8Tab[_r^v]
	}
	return
}

func CRC32(bs []byte) uint32 {
	return crc32.ChecksumIEEE(bs)
}

func CRC64(bs []byte) uint64 {
	return crc64.Checksum(bs, crc64.MakeTable(crc64.ECMA))
}

var seed = maphash.MakeSeed()

func Hash(key []byte) uint64 {
	return maphash.Bytes(seed, key)
}

func Rand(i int) (_r int) {
	if i > 0 {
		_r = rand.New(rand.NewSource(time.Now().UnixNano())).Intn(i)
	}
	return
}

func RandStrict(i int64) (_r int64, _err error) {
	if r, err := crand.Int(crand.Reader, big.NewInt(i)); err == nil {
		_r = r.Int64()
	} else {
		_err = err
	}
	return
}

func MatchString(pattern string, s string) bool {
	b, err := regexp.MatchString(pattern, s)
	if err != nil {
		b = false
	}
	return b
}

func StrToTimeFormat(s string) (t time.Time, err error) {
	l, _ := time.LoadLocation("Asia/Shanghai")
	t, err = time.ParseInLocation("2006-01-02 15:04:05", s, l)
	return
}

/***********************************************************/
var __inc uint64
var __pid = Int64ToBytes(int64(os.Getpid()))
var _rid, _ = RandStrict(1<<63 - 1)
var rids = Int64ToBytes(_rid)

func inc() uint64 {
	return atomic.AddUint64(&__inc, 1)
}

func RandId() (rid int64) {
	b := make([]byte, 24)
	copy(b[0:8], __pid)
	copy(b[8:], Int64ToBytes(time.Now().UnixNano()))
	copy(b[16:], rids)
	rid = int64(CRC32(b) & 0x7fffffff)
	rid = rid<<32 | int64(inc()&0x00000000ffffffff)
	return
}

func IsFileExist(path string) (_r bool) {
	if path != "" {
		_, err := os.Stat(path)
		_r = err == nil || os.IsExist(err)
	}
	return
}

func ReadFile(path string) (bs []byte, err error) {
	if r, er := mmap.Open(path); er == nil {
		defer r.Close()
		bs = make([]byte, r.Len())
		_, err = r.ReadAt(bs, 0)
	} else {
		err = er
	}
	return
}
