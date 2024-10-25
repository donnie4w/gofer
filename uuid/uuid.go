// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/uuid
// https://datatracker.ietf.org/doc/html/rfc4122
// https://en.wikipedia.org/wiki/Universally_unique_identifier

package uuid

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/donnie4w/gofer/base58"
	"hash/fnv"
	"hash/maphash"
	"math"
	mathrand "math/rand"
	"net"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"time"
)

type UUID [16]byte

func NewUUID() *UUID {
	var uuid UUID
	h64, seq := base.uniq()
	int64ToBytes(int64(h64), uuid[:8])
	int64ToBytes(seq, uuid[8:16])
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant
	return &uuid
}

func (u *UUID) String() (r string) {
	var dst [36]byte
	hex.Encode(dst[:], u[:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], u[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], u[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], u[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], u[10:])
	return string(dst[:])
}

func (u *UUID) Int64() int64 {
	return bytesToInt64(u[:8])
}

func (u *UUID) Int32() int32 {
	return int32(hash64(u[:]))
}

func (u *UUID) Bytes() []byte {
	return u[:]
}

// Version return version of UUID
func (u *UUID) Version() int {
	return int(u[6] >> 4)
}

func (u *UUID) Base58() []byte {
	return base58.EncodeForInt64(uint64(u.Int64()))
}

func (u *UUID) Equals(other *UUID) bool {
	if other == nil {
		return false
	}
	return u == other
}

func (u *UUID) Variant() byte {
	switch u[8] & 0xe0 {
	case 0x80:
		return 1 // RFC 4122
	case 0xc0:
		return 2 // Microsoft
	default:
		return 0 // NCS compatibility
	}
}

func Parse(uuidStr string) (*UUID, error) {
	if len(uuidStr) != 36 {
		return nil, errors.New("invalid UUID string length")
	}
	var uuid UUID
	strippedUUID := []byte(uuidStr[:8] + uuidStr[9:13] + uuidStr[14:18] + uuidStr[19:23] + uuidStr[24:])
	if _, err := hex.Decode(uuid[:], strippedUUID); err != nil {
		return nil, err
	}
	return &uuid, nil
}

var base = newBase()

type localbase struct {
	pid     [8]byte
	rid     [8]byte
	pidaddr [8]byte
	scale   byte
	seq     uint64
}

const maxSeq = 1<<60 - 1

func newBase() *localbase {
	var base localbase
	int64ToBytes(int64(hash64(getMachineID())), base.rid[:8])
	if _, err := rand.Read(base.pid[:]); err != nil {
		int64ToBytes(int64(os.Getpid()), base.pid[:])
	}
	int64ToBytes(int64(reflect.ValueOf(base.pid[:]).Pointer()), base.pidaddr[:])
	base.seq = uint64(mathrand.New(mathrand.NewSource(time.Now().UnixNano())).Intn(maxSeq))
	return &base
}

func (b *localbase) slice33(seq int64) []byte {
	var r [33]byte
	copy(r[0:8], b.pid[:])
	copy(r[8:16], b.rid[:])
	copy(r[16:24], b.pidaddr[:])
	int64ToBytes(int64(seq), r[24:32])
	r[32] = b.scale
	return r[:]
}

func (b *localbase) h128() [16]byte {
	return hash128(b.slice33(b.getSeq()))
}

func (b *localbase) uniq() (h64 uint64, seq int64) {
	seq = b.getSeq()
	h64 = hash64(b.slice33(seq))
	return
}

func (b *localbase) getSeq() int64 {
	seq := atomic.AddUint64(&b.seq, 1)
	if seq == math.MaxUint64 {
		b.scale++
	}
	return int64(seq)
}

var seed = maphash.MakeSeed()

func hash64(key []byte) uint64 {
	return maphash.Bytes(seed, key)
}

func hash128(input []byte) (h128 [16]byte) {
	hasher := fnv.New128a()
	hasher.Write(input)
	copy(h128[:], hasher.Sum(nil))
	return
}

func int64ToBytes(n int64, bs []byte) {
	for i := 0; i < 8; i++ {
		bs[i] = byte(n >> (8 * (7 - i)))
	}
}

func bytesToInt64(bs []byte) (_r int64) {
	if len(bs) >= 8 {
		for i := 0; i < 8; i++ {
			_r = _r | int64(bs[i])<<(8*(7-i))
		}
	} else {
		var bs8 [8]byte
		for i, b := range bs {
			bs8[i+8-len(bs)] = b
		}
		_r = bytesToInt64(bs8[:])
	}
	return
}

func getMachineID() []byte {
	sb := strings.Builder{}
	if interfaces, err := net.Interfaces(); err == nil {
		for _, iface := range interfaces {
			if iface.HardwareAddr != nil {
				sb.Write(iface.HardwareAddr)
			}
		}
	}

	if sb.Len() > 0 {
		return []byte(sb.String())
	}

	var bs [256]byte
	if _, err := rand.Read(bs[:]); err == nil {
		return bs[:]
	}

	ts := make([]byte, 0)
	var v uint64
	for range 1 << 7 {
		v += uint64(time.Now().UnixNano())
		m := make([]byte, 8)
		binary.LittleEndian.PutUint64(m, v)
		ts = append(ts, m...)
		m = make([]byte, 8)
		binary.LittleEndian.PutUint64(m, uint64(reflect.ValueOf(m).Pointer()))
		ts = append(ts, m...)
		time.Sleep(time.Duration(mathrand.Intn(100)) * time.Nanosecond)
	}
	return ts
}
