// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/cache

package cache

import (
	"fmt"
	"hash/maphash"
	"math"
	"sync"
	"sync/atomic"
)

// BloomFilter represents a bloom filter with a fixed number of bits and hash functions.
type BloomFilter struct {
	bits          [3][]uint64 // Bit array to store the presence information for each segment.
	hashCount     int         // Number of hash functions.
	seeds         []maphash.Seed
	mux           sync.Mutex
	numCount      atomic.Uint64
	arraySize     uint64
	expectedItems uint64
}

// NewBloomFilter creates a new bloom filter with the specified size, hash function count, and capacity.
func NewBloomFilter(expectedItems uint64, falsePositiveRate float64) (r *BloomFilter) {
	m := uint(-float64(expectedItems) * math.Log(falsePositiveRate) / (math.Log(2) * math.Log(2)))
	k := int(float64(m) / float64(expectedItems) * math.Log(2))
	if k == 0 {
		k = 1
	} else if k > 29 {
		k = 30
	}
	wordSize := uint64(64)
	words := (uint64(m) + wordSize - 1) / wordSize
	seeds := make([]maphash.Seed, k)
	for i := range seeds {
		seeds[i] = maphash.MakeSeed()
	}
	r = &BloomFilter{
		hashCount:     k,
		seeds:         seeds,
		arraySize:     words,
		expectedItems: expectedItems,
	}
	r.getBitIndex()
	return
}

func (bf *BloomFilter) String() string {
	return fmt.Sprintf("hashCount:%d,arraySize:%d,expectedItems:%d", bf.hashCount, bf.arraySize, bf.expectedItems)
}

// setBit sets the bit at the given index to 1.
func (bf *BloomFilter) setBit(index uint64) {
	wordIndex := index / 64
	bitIndex := index % 64
	id, _ := bf.getBitIndex()
	bf.bits[id][wordIndex] |= (uint64(1) << bitIndex)
}

// testBit checks if the bit at the given index is set (returns true if set, false otherwise).
func (bf *BloomFilter) testBit(index uint64) bool {
	defer recoverPanic(nil)
	wordIndex := index / 64
	bitIndex := index % 64
	id, prevId := bf.getBitIndex()
	offset := uint64(1) << bitIndex
	return (bf.bits[id][wordIndex]&offset) != 0 || (bf.bits[prevId][wordIndex]&offset) != 0
}

// Add adds an item (as byte slice) to the bloom filter using multiple hash functions.
func (bf *BloomFilter) Add(item []byte) {
	defer recoverPanic(nil)
	bf.numCount.Add(1)
	for i := 0; i < bf.hashCount; i++ {
		index := bf.hash(item, bf.seeds[i]) % (bf.arraySize * 64)
		bf.setBit(index)
	}
}

// Contains if an item (as byte slice) is possibly in the bloom filter.
func (bf *BloomFilter) Contains(item []byte) bool {
	defer recoverPanic(nil)
	for i := 0; i < bf.hashCount; i++ {
		index := bf.hash(item, bf.seeds[i]) % (bf.arraySize * 64)
		if !bf.testBit(index) {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) getBitIndex() (index, prevIndex int) {
	f := bf.numCount.Load() % (bf.expectedItems*3 + 1)
	rmIndex := 0
	if f < bf.expectedItems {
		index, rmIndex, prevIndex = 0, 1, 2
	} else if f < bf.expectedItems*2 {
		index, rmIndex, prevIndex = 1, 2, 0
	} else {
		index, rmIndex, prevIndex = 2, 0, 1
	}
	if bf.bits[index] == nil {
		bf.mux.Lock()
		if bf.bits[index] == nil {
			bf.bits[index] = make([]uint64, bf.arraySize)
		}
		if bf.bits[rmIndex] != nil {
			bf.bits[rmIndex] = nil
		}
		bf.mux.Unlock()
	}
	return
}

func (bf *BloomFilter) hash(item []byte, seed maphash.Seed) uint64 {
	return maphash.Bytes(seed, item)
}

func recoverPanic(err *error) {
	if r := recover(); r != nil {
		if err != nil {
			*err = fmt.Errorf("%v", r)
		}
	}
}
