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
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
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
	ib := FNVHash64([]byte("1234567789qwertyuiop"))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if FNVHash64([]byte("1234567789qwertyuiop")) != ib {
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
		//t.Log(RandUint(10))
		t.Log(RandUintCrypto(10))
	}
}

func BenchmarkRandUint(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			RandUint(10)
		}
	})
}

// 测试并发调用 Hash64
func TestMaphashRace(t *testing.T) {
	//  强制Go使用所有CPU核心（关键）
	runtime.GOMAXPROCS(runtime.NumCPU())

	//  准备128字节随机数据（放大雪崩效应）
	key := make([]byte, 128)
	for i := range key {
		key[i] = byte(i%256 + 1) // 避免0值，放大Seed错误的影响
	}

	//  单协程计算预期值（确保初始值正确）
	expected := Hash64(key)
	t.Logf("预期hash值：%d", expected)

	//  原子变量统计错误数（避免竞态）
	var errCount uint64
	var wg sync.WaitGroup

	//  启动与CPU核心数匹配的协程（核心：让每个核心都跑满，并行读取Seed）
	cpuNum := runtime.NumCPU()
	goroutineNum := cpuNum * 2000 // 每个核心跑2000个协程，制造并行压力
	for i := 0; i < goroutineNum; i++ {
		wg.Add(1)
		// 闭包捕获i，避免协程复用（关键）
		go func(idx int) {
			defer wg.Done()
			// 内层循环加随机休眠，强制触发协程切换（核心）
			for j := 0; j < 100000; j++ {
				// 随机休眠1ns，强制CPU切换协程，制造Seed读取中断
				if j%100 == 0 {
					runtime.Gosched() // 主动让出CPU，放大调度中断概率
				}

				actual := Hash64(key)
				if actual != expected {
					atomic.AddUint64(&errCount, 1)
					// 只打印一次错误，避免刷屏
					if atomic.LoadUint64(&errCount) == 1 {
						t.Errorf("协程%d：hash值不一致！预期=%d，实际=%d", idx, expected, actual)
						// 触发错误后直接退出，无需继续测试
						os.Exit(1)
					}
				}
			}
		}(i)
	}

	// 启动CPU压力协程，模拟生产环境负载
	go func() {
		for {
			_ = 1 + 1 // 空循环占用CPU，制造调度压力
		}
	}()

	wg.Wait()
	t.Logf("测试完成，总错误数：%d", errCount)
}
