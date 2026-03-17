package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"regexp"
	"time"
)

func ExecTime(fn func()) float64 {
	start := time.Now()
	fn()
	tc := float64(time.Since(start).Nanoseconds())
	return tc / 1e6
}

func ExecTimeWithError(fn func() error) (float64, error) {
	start := time.Now()
	err := fn()
	tc := float64(time.Since(start).Nanoseconds())
	return tc / 1e6, err
}

func Encoder(data interface{}) []byte {
	if data == nil {
		return nil
	}
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(data)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func Decoder(data []byte, v interface{}) {
	if data == nil {
		return
	}
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(v)
	if err != nil {
		panic(err)
	}
}

func init() {
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})
}

const (
	c1 = 0xcc9e2d51
	c2 = 0x1b873593
	c3 = 0x85ebca6b
	c4 = 0xc2b2ae35
	r1 = 15
	r2 = 13
	m  = 5
	n  = 0xe6546b64
)

var (
	Seed = uint32(1)
)

func Murmur3(key []byte) (hash uint32) {
	hash = Seed
	iByte := 0
	for ; iByte+4 <= len(key); iByte += 4 {
		k := uint32(key[iByte]) | uint32(key[iByte+1])<<8 | uint32(key[iByte+2])<<16 | uint32(key[iByte+3])<<24
		k *= c1
		k = (k << r1) | (k >> (32 - r1))
		k *= c2
		hash ^= k
		hash = (hash << r2) | (hash >> (32 - r2))
		hash = hash*m + n
	}

	var remainingBytes uint32
	switch len(key) - iByte {
	case 3:
		remainingBytes += uint32(key[iByte+2]) << 16
		fallthrough
	case 2:
		remainingBytes += uint32(key[iByte+1]) << 8
		fallthrough
	case 1:
		remainingBytes += uint32(key[iByte])
		remainingBytes *= c1
		remainingBytes = (remainingBytes << r1) | (remainingBytes >> (32 - r1))
		remainingBytes = remainingBytes * c2
		hash ^= remainingBytes
	}

	hash ^= uint32(len(key))
	hash ^= hash >> 16
	hash *= c3
	hash ^= hash >> 13
	hash *= c4
	hash ^= hash >> 16

	return
}

func StringToHash(s string) uint32 {
	return Murmur3([]byte(s))
}

func Uint32Comparator(a, b interface{}) int {
	UintA := a.(uint32)
	UintB := b.(uint32)
	if UintA < UintB {
		return -1
	} else if UintA > UintB {
		return 1
	}
	return 0
}

func Uint32ToBtye(u uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, u)
	return b
}

func DeleteArray(arr []byte, index int) []byte {
	return append(arr[:index], arr[index+1:]...)
}

func RemovePunctuation(s string) string {
	reg := regexp.MustCompile(`\p{P}+`)
	return reg.ReplaceAllString(s, "")
}

func RemoveSpace(s string) string {
	reg := regexp.MustCompile(`\s+`)
	return reg.ReplaceAllString(s, "")
}
