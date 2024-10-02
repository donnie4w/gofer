// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
// github.com/donnie4w/gofer/uuid

package uuid

import (
	"fmt"
	"github.com/donnie4w/gofer/hashmap"
	"testing"
)

func Benchmark_NewUUID(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for range 10 {
				//NewUUID().Int32()
				NewUUID().Int64()
				//NewUUID().Base58()
				//NewUUID().String()
			}
		}
	})
}

func Benchmark_UUID_UNI(b *testing.B) {
	m := hashmap.NewMap[any, byte]()
	for range b.N {
		if id := NewUUID().Int64(); m.Has(id) {
			fmt.Println(id)
			break
		} else {
			m.Put(id, 1)
		}
	}
}

func Test_UUID_UNI(t *testing.T) {
	m := hashmap.NewMap[any, byte]()
	for range 1 << 24 {
		if id := NewUUID().Int64(); m.Has(id) {
			fmt.Println(id)
			break
		} else {
			m.Put(id, 1)
		}
	}
}

func Test_NewUUID(t *testing.T) {
	for range 10 {
		fmt.Println(NewUUID().String())
	}
}
