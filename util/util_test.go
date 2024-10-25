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
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/golang/snappy"
)

func TestDataToBytes(t *testing.T) {
	arr := []int64{1 << 60, 2 << 60, 3, 4}
	bs := IntArrayToBytes(arr)
	fmt.Println(bs)
	arr2 := BytesToIntArray(bs)
	fmt.Println(arr2)
}

func TestZlib(t *testing.T) {
	in := []byte("123456789")
	bs, err := Zlib(in)
	fmt.Println(err)
	fmt.Println(string(bs))
	bs, err = UnZlib(bs)
	fmt.Println(err)
	fmt.Println(string(bs))
}

func Benchmark_md5(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Md5Str("1234567890qwertyuiop")
	}
}

func Benchmark_crc32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CRC32([]byte("1234567890qwertyuiop1234567890qwertyuiop1234567890qwertyuiop1234567890qwertyuiop"))
	}
}

func Benchmark_crc64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CRC64([]byte("1234567890qwertyuiop1234567890qwertyuiop1234567890qwertyuiop1234567890qwertyuiop"))
	}
}

func Benchmark_bs2int(b *testing.B) {
	b.StopTimer()
	bs := Int64ToBytes(99)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		BytesToInt64(bs)
	}
}

func Benchmark_int64Tobs(b *testing.B) {
	b.StopTimer()
	t := time.Now().UnixNano()
	bs := Int64ToBytes(t)
	b.StartTimer()
	var r int64
	for i := 0; i < b.N; i++ {
		r = BytesToInt64(bs)
	}
	fmt.Println(t == r)
}

func Test_int64Tobs(t *testing.T) {
	for i := int64(1 << 1); i < 1000; i++ {
		if bs := Int64ToBytes(i); i != BytesToInt64(bs) {
			panic("err >>" + fmt.Sprint(i))
		}
	}
	fmt.Println(BytesToInt64([]byte{0, 1}))
	fmt.Println("ok")
}

func Test_int32Tbs(t *testing.T) {
	for i := int32(0); i < 1<<30; i++ {
		if bs := Int32ToBytes(i); i != BytesToInt32(bs) {
			panic("err >>" + fmt.Sprint(i))
		}
	}
	fmt.Println("ok")
}

func Benchmark_maphash(b *testing.B) {
	ib := Hash64([]byte("1234567789qwertyuiop"))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if Hash64([]byte("1234567789qwertyuiop")) != ib {
				panic("err")
			}
		}
	})
}

func Benchmark_czlib(b *testing.B) {
	b.StopTimer()
	bs, _ := ReadFile("gob.go")
	fmt.Println("len(bs)>>", len(bs))
	var r []byte
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		r, _ = Zlib(bs)
	}
	fmt.Println("len(r)>>", len(r))
}

func Benchmark_snappy(b *testing.B) {
	b.StopTimer()
	bs, _ := ReadFile("gob.go")
	fmt.Println("len(bs)>>", len(bs))
	var dst []byte
	dst = snappy.Encode(nil, bs)
	var bs2 []byte
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		bs2, _ = snappy.Decode(nil, dst)
	}
	fmt.Println("len(r)>>", len(bs2))
}

func TestIntByte(t *testing.T) {
	b1 := Int16ToBytes(1<<15 - 1)
	fmt.Println("b >>", BytesToInt16(b1))
	b2 := Int32ToBytes(1<<31 - 1)
	fmt.Println("b >>", BytesToInt32(b2))
	b3 := Int64ToBytes(1<<63 - 1)
	fmt.Println("b >>", BytesToInt64(b3))
	var i float32 = 0.01
	var byt bytes.Buffer
	binary.Write(&byt, binary.BigEndian, i)
	binary.Read(&byt, binary.BigEndian, &i)
	fmt.Println(i)
	j, _ := strconv.ParseFloat("0.11", 32)
	fmt.Println(float32(j))
}

func Test_Gzip(t *testing.T) {
	buf, err := Gzip([]byte("hello123"))
	buf, err = UnGzip(buf.Bytes())
	fmt.Println(err)
	fmt.Println(string(buf.Bytes()))
}

func TestCrc(t *testing.T) {
	fmt.Printf("CRC-8: %X\n", CRC8([]byte("12")))
}

func TestRandUint(t *testing.T) {
	for range 10 {
		t.Log(RandUint(10))
	}
}

func BenchmarkRandUint(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			RandUint(10)
		}
	})
}
