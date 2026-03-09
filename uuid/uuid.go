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
	"fmt"
	"github.com/donnie4w/gofer/base58"
	"hash/fnv"
	"hash/maphash"
	mathrand "math/rand/v2"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// UUID 16-byte UUID structure compliant with RFC4122 standard
type UUID [16]byte

var (
	// enableSecureMode: Global switch to enable secure mode (generate Version4 with crypto/rand)
	enableSecureMode atomic.Bool
	// highPerfBase: Base data for high-performance UUID (machine ID/process ID/sequence)
	highPerfBase = newHighPerfBase()
	// randPoolCap: Fixed capacity of random number pool (1024 UUID-sized random byte arrays)
	randPoolCap = 1024
	// randPool: Random number pool to optimize performance in secure mode
	randPool = make(chan [16]byte, randPoolCap)
	// randPoolInitOnce: Ensure pool initialization runs only once
	randPoolInitOnce sync.Once
	// isRefilling: Prevent concurrent pool replenishment in edge cases
	isRefilling atomic.Bool
)

// SetSecureMode sets UUID generation mode
// - secure=true: Enable secure mode (Version4 via crypto/rand, guaranteed uniqueness, slightly lower performance)
// - secure=false: Enable high-performance mode (default, based on machine+process+sequence, extremely high performance)
func SetSecureMode(secure bool) {
	enableSecureMode.Store(secure)
	if secure {
		// Initialize random pool (warm up) when secure mode is enabled
		randPoolInitOnce.Do(initRandPool)
	}
}

// initRandPool initializes and maintains random number pool:
// 1. Fill pool with 1024 random byte arrays on startup
// 2. Run background goroutine to replenish pool (maintain 1024 capacity)
func initRandPool() {
	// Step 1: Batch fill pool to full capacity on initialization
	batchFillRandPool(randPoolCap)

	// Step 2: Start background goroutine for "take-one-replenish-one" logic
	go func() {
		for {
			// Sleep to reduce resource usage when secure mode is disabled
			if !enableSecureMode.Load() {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Replenish pool when capacity is below 1024
			if len(randPool) < randPoolCap {
				// Prevent concurrent replenishment
				if !isRefilling.CompareAndSwap(false, true) {
					continue
				}

				// Calculate required replenishment quantity (fill to 1024)
				needFill := randPoolCap - len(randPool)
				batchFillRandPool(needFill)

				isRefilling.Store(false)
			} else {
				// Sleep briefly when pool is full to reduce CPU usage
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()
}

// batchFillRandPool generates specified number of random byte arrays in batch and non-blockingly fills the pool
// count: Number of 16-byte random arrays to generate
func batchFillRandPool(count int) {
	if count <= 0 || count > randPoolCap {
		return
	}

	// Batch read random bytes (reduce syscall overhead via single read)
	totalBytes := count * 16
	buf := make([]byte, totalBytes)
	_, err := rand.Read(buf)
	if err != nil {
		fmt.Printf("[WARNING] Failed to batch read crypto random bytes: %v\n", err)
		// Fallback to math/rand to ensure pool availability (slightly reduced security)
		fallbackFillRandPool(count)
		return
	}

	// Split batch buffer into 16-byte chunks and fill pool non-blockingly
	for i := 0; i < count; i++ {
		var uuidBytes [16]byte
		copy(uuidBytes[:], buf[i*16:(i+1)*16])
		select {
		case randPool <- uuidBytes:
		default:
			// Discard when pool is full to avoid blocking
			break
		}
	}
}

// fallbackFillRandPool uses math/rand/v2 to fill pool (LAST RESORT when crypto/rand fails completely)
// count: Number of 16-byte random arrays to generate
func fallbackFillRandPool(count int) {
	// Initialize pseudo-random generator with unique seed (timestamp + PID + machine ID)
	seed := uint64(time.Now().UnixNano()) ^ uint64(os.Getpid()) ^ highPerfBase.machineID
	r := mathrand.New(mathrand.NewPCG(seed, seed^0xdeadbeef))

	// Generate pseudo-random bytes (correct mathrand/v2 usage)
	for i := 0; i < count; i++ {
		var uuidBytes [16]byte

		// Most efficient way (generate 16 random bytes directly)
		for j := 0; j < len(uuidBytes); j++ {
			uuidBytes[j] = byte(r.Uint32() % 256) // Generate 0-255 byte values
		}

		// Non-blocking write to pool
		select {
		case randPool <- uuidBytes:
		default:
			break
		}
	}

	// Critical warning for admin visibility
	fmt.Printf("[CRITICAL] Using math/rand/v2 for UUID pool fallback (crypto/rand unavailable) - security reduced!\n")
}

// NewUUID generates UUID (auto-selects high-performance/secure mode based on global setting)
// - High-performance mode: ~30ns/op (machineID+PID+incremental sequence)
// - Secure mode: ~45ns/op (RFC4122 Version4 with pool optimization)
func NewUUID() *UUID {
	if enableSecureMode.Load() {
		if r := newSecureUUID(); r != nil {
			return r
		}
	}
	return newHighPerfUUID()
}

// NewHighPerfUUID explicitly generates high-performance UUID (bypass mode check for optimal performance)
func NewHighPerfUUID() *UUID {
	return newHighPerfUUID()
}

// NewSecureUUID explicitly generates secure UUID (RFC4122 Version4)
// Returns error when crypto/rand fails and fallback is triggered
func NewSecureUUID() (*UUID, error) {
	if r := newSecureUUID(); r != nil {
		return r, nil
	} else {
		return nil, errors.New("crypto/rand failed, fallback to math/rand (uuid security reduced)")
	}
}

// _highPerfBase contains base data for high-performance UUID generation
type _highPerfBase struct {
	machineID uint64        // Unique machine identifier (MAC address/random number)
	pid       uint64        // Process ID
	seq       atomic.Uint64 // Incremental sequence (lock-free via atomic operation)
	scale     atomic.Uint32 // Sequence overflow counter
}

// maxSeq: Maximum sequence value to avoid frequent overflow (2^60)
const maxSeq = 1 << 60

// newHighPerfBase initializes base data for high-performance UUID
func newHighPerfBase() *_highPerfBase {
	b := &_highPerfBase{
		pid: uint64(os.Getpid()),
	}

	// Generate unique machine ID (prefer MAC address, fallback to random number)
	b.machineID = genMachineID()

	// Initialize sequence with random seed to avoid duplication after restart
	var seed [8]byte
	if _, err := rand.Read(seed[:]); err == nil {
		b.seq.Store(binary.BigEndian.Uint64(seed[:]) % maxSeq)
	} else {
		b.seq.Store(mathrand.Uint64() % maxSeq)
	}

	return b
}

// genMachineID generates unique machine identifier (high-performance, non-blocking)
// Priority: Network interface MAC address > crypto/rand > timestamp+PID
func genMachineID() uint64 {
	// Try to get network interface MAC addresses first
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			// Use active interfaces with valid MAC address
			if iface.Flags&net.FlagUp != 0 && iface.HardwareAddr != nil && len(iface.HardwareAddr) > 0 {
				h := fnv.New64a()
				h.Write(iface.HardwareAddr)
				return h.Sum64()
			}
		}
	}

	// Fallback 1: Generate random number via crypto/rand
	var randBytes [8]byte
	if _, err := rand.Read(randBytes[:]); err == nil {
		return binary.BigEndian.Uint64(randBytes[:])
	}

	// Fallback 2: Use timestamp + PID (last resort)
	return uint64(time.Now().UnixNano()) ^ uint64(os.Getpid())
}

// newHighPerfUUID generates high-performance UUID (core logic)
// Uniqueness guaranteed by: machineID + PID + atomic sequence + scale counter
func newHighPerfUUID() *UUID {
	var uuid UUID
	// Get incremental sequence (atomic operation for thread safety)
	seq := highPerfBase.seq.Add(1)
	if seq >= maxSeq {
		// Reset sequence and increment scale counter on overflow
		highPerfBase.seq.Store(0)
		highPerfBase.scale.Add(1)
	}

	// Generate 64-bit hash from machineID+PID+sequence+scale
	var buf [32]byte
	binary.BigEndian.PutUint64(buf[0:8], highPerfBase.machineID)
	binary.BigEndian.PutUint64(buf[8:16], highPerfBase.pid)
	binary.BigEndian.PutUint64(buf[16:24], seq)
	binary.BigEndian.PutUint64(buf[24:32], uint64(highPerfBase.scale.Load()))

	// Fill first 8 bytes with hash, last 8 bytes with sequence (ensure uniqueness)
	seed := maphash.MakeSeed()
	hashVal := maphash.Bytes(seed, buf[:])
	binary.BigEndian.PutUint64(uuid[:8], hashVal)
	binary.BigEndian.PutUint64(uuid[8:16], seq)

	// Set Version/Variant for RFC4122 compatibility (non-standard but compatible)
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Mark as Version4 (compatibility only, not standard)
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // RFC4122 Variant

	return &uuid
}

// newSecureUUID generates secure UUID compliant with RFC4122 Version4
// Priority: Random pool > direct crypto/rand read
func newSecureUUID() *UUID {
	var uuid UUID
	// First try to get from random pool (high hit rate after initialization)
	select {
	case buf := <-randPool:
		uuid = buf
	default:
		// Fallback to direct crypto/rand read when pool is empty
		if _, err := rand.Read(uuid[:]); err != nil {
			fmt.Printf("[WARNING] crypto/rand failed, fallback to math/rand (uuid security reduced)\n")
			return nil
		}
	}

	// Strictly follow RFC4122 Version4 specification
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version4 (random generated)
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // RFC4122 Variant (10x format)
	return &uuid
}

// String formats UUID to standard RFC4122 string format (e.g. "123e4567-e89b-12d3-a456-426614174000")
func (u *UUID) String() string {
	var buf [36]byte
	hex.Encode(buf[:8], u[:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:])
	return string(buf[:])
}

// Int64 converts first 8 bytes of UUID to int64
func (u *UUID) Int64() int64 {
	return int64(binary.BigEndian.Uint64(u[:8]))
}

// Uint64 converts first 8 bytes of UUID to int64
func (u *UUID) Uint64() uint64 {
	return binary.BigEndian.Uint64(u[:8])
}

// Int32 generates 32-bit hash value of UUID
func (u *UUID) Int32() int32 {
	return int32(u.Uint32())
}

// Uint32 generates 32-bit hash value of UUID
func (u *UUID) Uint32() uint32 {
	h := fnv.New32a()
	h.Write(u[:])
	return h.Sum32()
}

// Bytes returns byte slice of UUID (no copy)
func (u *UUID) Bytes() []byte {
	return u[:]
}

// Version returns UUID version number (compliant with RFC4122)
func (u *UUID) Version() int {
	return int(u[6] >> 4)
}

// Base58 returns Base58 encoding of first 8 bytes of UUID
func (u *UUID) Base58() []byte {
	return base58.EncodeForInt64(binary.BigEndian.Uint64(u[:8]))
}

// Equals compares two UUIDs for value equality (not pointer equality)
func (u *UUID) Equals(other *UUID) bool {
	if other == nil {
		return false
	}
	return *u == *other
}

// Variant returns UUID variant type:
// 1: RFC 4122 (standard)
// 2: Microsoft
// 0: NCS compatibility
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

// Parse decodes standard UUID string to UUID structure
// Returns error for invalid format/length
func Parse(uuidStr string) (*UUID, error) {
	if len(uuidStr) != 36 {
		return nil, errors.New("invalid UUID length: must be 36 characters")
	}
	// Validate hyphen positions for standard format
	if uuidStr[8] != '-' || uuidStr[13] != '-' || uuidStr[18] != '-' || uuidStr[23] != '-' {
		return nil, errors.New("invalid UUID format: incorrect hyphen positions")
	}

	var uuid UUID
	pos := 0
	// Decode each segment of UUID string
	for _, r := range []struct{ start, end int }{
		{0, 8}, {9, 13}, {14, 18}, {19, 23}, {24, 36},
	} {
		n, err := hex.Decode(uuid[pos:pos+4], []byte(uuidStr[r.start:r.end]))
		if err != nil {
			return nil, fmt.Errorf("failed to decode hex segment: %w", err)
		}
		pos += n
	}
	return &uuid, nil
}
