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

func MD5(s string) string {
	m := md5.New()
	m.Write([]byte(s))
	return strings.ToUpper(hex.EncodeToString(m.Sum(nil)))
}

func SHA1(s string) string {
	m := sha1.New()
	m.Write([]byte(s))
	return strings.ToUpper(hex.EncodeToString(m.Sum(nil)))
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
var __pid = int64(os.Getpid())
var _dir, _ = os.Getwd()
var __dir = []byte(_dir)
func inc() uint64 {
	return atomic.AddUint64(&__inc, 1)
}

func RandId() (rid int64) {
	b := make([]byte, 16+len(__dir))
	copy(b[0:8], Int64ToBytes(__pid))
	copy(b[8:], Int64ToBytes(time.Now().UnixNano()))
	copy(b[16:], __dir)
	rid = int64(CRC32(b)&0x7fffffff)
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
