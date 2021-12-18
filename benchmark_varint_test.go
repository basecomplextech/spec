package spec

import (
	"encoding/binary"
	"math"
	"testing"
)

// get

func BenchmarkReverseUvarint32(b *testing.B) {
	buf := make([]byte, maxReverseVarintLen64)
	putReverseUvarint(buf, math.MaxUint32)
	b.SetBytes(8)

	for i := 0; i < b.N; i++ {
		v, n := reverseUvarint(buf)
		if v != math.MaxUint32 {
			b.Fatal()
		}
		if n != maxReverseVarintLen32 {
			b.Fatal()
		}
	}
}

func BenchmarkReverseUvarint64(b *testing.B) {
	buf := make([]byte, maxReverseVarintLen64)
	putReverseUvarint(buf, math.MaxUint64)
	b.SetBytes(8)

	for i := 0; i < b.N; i++ {
		v, n := reverseUvarint(buf)
		if v != math.MaxUint64 {
			b.Fatal()
		}
		if n != maxReverseVarintLen64 {
			b.Fatal()
		}
	}
}

// put

func BenchmarkPutReverseUvarint32(b *testing.B) {
	buf := make([]byte, maxReverseVarintLen32)
	b.SetBytes(4)

	for i := 0; i < b.N; i++ {
		n := putReverseUvarint(buf, math.MaxUint32)
		if n != maxReverseVarintLen32 {
			b.Fatal()
		}
	}
}

func BenchmarkPutReverseUvarint64(b *testing.B) {
	buf := make([]byte, maxReverseVarintLen64)
	b.SetBytes(8)

	for i := 0; i < b.N; i++ {
		n := putReverseUvarint(buf, math.MaxUint64)
		if n != maxReverseVarintLen64 {
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
