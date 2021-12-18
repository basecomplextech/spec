package spec

import (
	"encoding/binary"
	"math"
)

func writeType(b []byte, type_ Type) []byte {
	return append(b, byte(type_))
}

func writeInt8(b []byte, v int8) []byte {
	return append(b, uint8(v))
}

func writeInt16(b []byte, v int16) []byte {
	p := [maxVarintLen16]byte{}
	off := putReverseVarint(p[:], int64(v))
	return append(b, p[off:]...)
}

func writeInt32(b []byte, v int32) []byte {
	p := [maxVarintLen32]byte{}
	off := putReverseVarint(p[:], int64(v))
	return append(b, p[off:]...)
}

func writeInt64(b []byte, v int64) []byte {
	p := [maxVarintLen64]byte{}
	off := putReverseVarint(p[:], v)
	return append(b, p[off:]...)
}

func writeUInt8(b []byte, v uint8) []byte {
	return append(b, v)
}

func writeUInt16(b []byte, v uint16) []byte {
	p := [maxVarintLen16]byte{}
	off := putReverseUvarint(p[:], uint64(v))
	return append(b, p[off:]...)
}

func writeUInt32(b []byte, v uint32) []byte {
	p := [maxVarintLen32]byte{}
	off := putReverseUvarint(p[:], uint64(v))
	return append(b, p[off:]...)
}

func writeUInt64(b []byte, v uint64) []byte {
	p := [maxVarintLen64]byte{}
	off := putReverseUvarint(p[:], v)
	return append(b, p[off:]...)
}

func writeFloat32(b []byte, v float32) []byte {
	p := [4]byte{}
	binary.BigEndian.PutUint32(p[:], math.Float32bits(v))
	return append(b, p[:]...)
}

func writeFloat64(b []byte, v float64) []byte {
	p := [8]byte{}
	binary.BigEndian.PutUint64(p[:], math.Float64bits(v))
	return append(b, p[:]...)
}

// bytes

func writeBytes(b []byte, v []byte) ([]byte, uint32) {
	b = append(b, v...)
	size := uint32(len(v))
	return b, size
}

func writeBytesSize(b []byte, size uint32) []byte {
	p := [4]byte{}
	binary.BigEndian.PutUint32(p[:], uint32(size))
	return append(b, p[:]...)
}

// string

func writeString(b []byte, s string) ([]byte, uint32) {
	b = append(b, s...)
	size := uint32(len(s))
	return b, size
}

func writeStringZero(b []byte) []byte {
	return append(b, 0)
}

func writeStringSize(b []byte, size uint32) []byte {
	p := [4]byte{}
	binary.BigEndian.PutUint32(p[:], uint32(size))
	return append(b, p[:]...)
}

// list

func writeListTable(b []byte, table []listElement) ([]byte, uint32) {
	size := len(table) * listElementSize
	b, p := writeAlloc(b, size)

	off := 0
	for _, elem := range table {
		q := p[off : off+listElementSize]
		binary.BigEndian.PutUint32(q, elem.offset)
		off += listElementSize
	}

	return b, uint32(size)
}

func writeListTableSize(b []byte, size uint32) []byte {
	p := [4]byte{}
	binary.BigEndian.PutUint32(p[:], size)
	return append(b, p[:]...)
}

func writeListDataSize(b []byte, size uint32) []byte {
	p := [4]byte{}
	binary.BigEndian.PutUint32(p[:], size)
	return append(b, p[:]...)
}

// message

func writeMessageTable(b []byte, table []messageField) ([]byte, uint32) {
	size := len(table) * messageFieldSize
	b, p := writeAlloc(b, size)

	off := 0
	for _, field := range table {
		q := p[off : off+messageFieldSize]

		binary.BigEndian.PutUint16(q, field.tag)
		binary.BigEndian.PutUint32(q[2:], field.offset)

		off += messageFieldSize
	}

	return b, uint32(size)
}

func writeMessageTableSize(b []byte, size uint32) []byte {
	p := [4]byte{}
	binary.BigEndian.PutUint32(p[:], size)
	return append(b, p[:]...)
}

func writeMessageDataSize(b []byte, size uint32) []byte {
	p := [4]byte{}
	binary.BigEndian.PutUint32(p[:], size)
	return append(b, p[:]...)
}

// private

// writeAlloc grows a buffer by n bytes and returns a new buffer and an allocated segment.
func writeAlloc(b []byte, n int) ([]byte, []byte) {
	cp := cap(b)
	ln := len(b)

	// alloc
	rem := cp - ln
	if rem < n {
		size := (cp * 2) + n
		buf := make([]byte, ln, size)
		copy(buf, b)
		b = buf
	}

	// return
	size := ln + n
	b = b[:size]
	p := b[ln:size]
	return b, p
}
