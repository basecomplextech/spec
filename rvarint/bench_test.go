package rvarint

import (
	"encoding/binary"
	"math"
	"testing"
)

// Decode

func BenchmarkUint32(b *testing.B) {
	buf := make([]byte, MaxLen32)
	PutUint64(buf, math.MaxUint32)
	b.SetBytes(8)

	for i := 0; i < b.N; i++ {
		v, n := Uint32(buf)
		if v != math.MaxUint32 {
			b.Fatal()
		}
		if n <= 0 {
			b.Fatal()
		}
	}
}

func BenchmarkUint64(b *testing.B) {
	buf := make([]byte, MaxLen64)
	PutUint64(buf, math.MaxUint64)
	b.SetBytes(8)

	for i := 0; i < b.N; i++ {
		v, n := Uint64(buf)
		if v != math.MaxUint64 {
			b.Fatal()
		}
		if n <= 0 {
			b.Fatal()
		}
	}
}

// Encode

func BenchmarkPutUint32(b *testing.B) {
	buf := make([]byte, MaxLen32)
	b.SetBytes(4)

	for i := 0; i < b.N; i++ {
		n := PutUint64(buf, math.MaxUint32)
		if n <= 0 {
			b.Fatal()
		}
	}
}

func BenchmarkPutUint64(b *testing.B) {
	buf := make([]byte, MaxLen64)
	b.SetBytes(8)

	for i := 0; i < b.N; i++ {
		n := PutUint64(buf, math.MaxUint64)
		if n <= 0 {
			b.Fatal()
		}
	}
}

// Standard varint

func BenchmarkUvarint64(b *testing.B) {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(buf, math.MaxUint64)
	b.SetBytes(8)

	for i := 0; i < b.N; i++ {
		v, n := binary.Uvarint(buf)
		if v != math.MaxUint64 {
			b.Fatal()
		}
		if n != binary.MaxVarintLen64 {
			b.Fatal()
		}
	}
}

func BenchmarkPutUvarint64(b *testing.B) {
	buf := make([]byte, binary.MaxVarintLen64)
	b.SetBytes(8)

	for i := 0; i < b.N; i++ {
		n := binary.PutUvarint(buf, math.MaxUint64)
		if n != binary.MaxVarintLen64 {
			b.Fatal()
		}
	}
}

// Standard big endian

func BenchmarkBigEndianUint64(b *testing.B) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, math.MaxUint64)
	b.SetBytes(8)

	for i := 0; i < b.N; i++ {
		v := binary.BigEndian.Uint64(buf)
		if v != math.MaxUint64 {
			b.Fatal()
		}
	}
}

func BenchmarkPutBigEndianUint64(b *testing.B) {
	buf := make([]byte, 8)
	b.SetBytes(8)

	for i := 0; i < b.N; i++ {
		binary.BigEndian.PutUint64(buf, math.MaxUint64)
	}
}
