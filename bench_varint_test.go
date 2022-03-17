package spec

import (
	"encoding/binary"
	"math"
	"testing"
)

// read

func BenchmarkReadReverseUvarint32(b *testing.B) {
	buf := make([]byte, maxVarintLen32)
	writeReverseUvarint(buf, math.MaxUint32)
	b.SetBytes(8)

	for i := 0; i < b.N; i++ {
		v, n := readReverseUvarint(buf)
		if v != math.MaxUint32 {
			b.Fatal()
		}
		if n <= 0 {
			b.Fatal()
		}
	}
}

func BenchmarkReadReverseUvarint64(b *testing.B) {
	buf := make([]byte, maxVarintLen64)
	writeReverseUvarint(buf, math.MaxUint64)
	b.SetBytes(8)

	for i := 0; i < b.N; i++ {
		v, n := readReverseUvarint(buf)
		if v != math.MaxUint64 {
			b.Fatal()
		}
		if n <= 0 {
			b.Fatal()
		}
	}
}

// write

func BenchmarkWriteReverseUvarint32(b *testing.B) {
	buf := make([]byte, maxVarintLen32)
	b.SetBytes(4)

	for i := 0; i < b.N; i++ {
		n := writeReverseUvarint(buf, math.MaxUint32)
		if n <= 0 {
			b.Fatal()
		}
	}
}

func BenchmarkWriteReverseUvarint64(b *testing.B) {
	buf := make([]byte, maxVarintLen64)
	b.SetBytes(8)

	for i := 0; i < b.N; i++ {
		n := writeReverseUvarint(buf, math.MaxUint64)
		if n <= 0 {
			b.Fatal()
		}
	}
}

// standard varint

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

// standard big endian

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
