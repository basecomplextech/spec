package protocol

import (
	"encoding/binary"
	"math"
)

type writeBuffer []byte

func (b *writeBuffer) offset() int {
	return len(*b)
}

// type and size

func (b *writeBuffer) writeType(type_ Type) {
	p := b._grow(1)
	p[0] = byte(type_)
}

func (b *writeBuffer) writeSize(size uint32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, uint32(size))
}

// values

func (b *writeBuffer) writeByte(v byte) {
	p := b._grow(1)
	p[0] = v
}

func (b *writeBuffer) writeInt8(v int8) {
	p := b._grow(1)
	p[0] = byte(v)
}

func (b *writeBuffer) writeInt16(v int16) {
	p := b._grow(2)
	binary.BigEndian.PutUint16(p, uint16(v))
}

func (b *writeBuffer) writeInt32(v int32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, uint32(v))
}

func (b *writeBuffer) writeInt64(v int64) {
	p := b._grow(8)
	binary.BigEndian.PutUint64(p, uint64(v))
}

func (b *writeBuffer) writeUInt8(v uint8) {
	p := b._grow(1)
	p[0] = v
}

func (b *writeBuffer) writeUInt16(v uint16) {
	p := b._grow(2)
	binary.BigEndian.PutUint16(p, v)
}

func (b *writeBuffer) writeUInt32(v uint32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, v)
}

func (b *writeBuffer) writeUInt64(v uint64) {
	p := b._grow(8)
	binary.BigEndian.PutUint64(p, v)
}

func (b *writeBuffer) writeFloat32(v float32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, math.Float32bits(v))
}

func (b *writeBuffer) writeFloat64(v float64) {
	p := b._grow(8)
	binary.BigEndian.PutUint64(p, math.Float64bits(v))
}

// bytes

func (b *writeBuffer) writeBytes(v []byte) {
	p := b._grow(len(v))
	copy(p, v)
}

// string

func (b *writeBuffer) writeString(s string) {
	size := len(s)
	p := b._grow(size)
	copy(p, s)
}

func (b *writeBuffer) writeStringZero() {
	p := b._grow(1)
	p[0] = 0
}

// list

func (b *writeBuffer) writeElements(list []element) {
	size := len(list) * elementSize
	p := b._grow(size)

	for i, elem := range list {
		off := i * elementSize
		q := p[off : off+elementSize]

		binary.BigEndian.PutUint32(q, elem.offset)
	}
}

func (b *writeBuffer) writeElementCount(count uint32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, count)
}

// struct

func (b *writeBuffer) writeFields(table []field) {
	size := len(table) * fieldSize
	p := b._grow(size)

	for i, field := range table {
		off := i * fieldSize
		q := p[off : off+fieldSize]

		binary.BigEndian.PutUint16(q, field.tag)
		binary.BigEndian.PutUint32(q[2:], field.offset)
	}
}

func (b *writeBuffer) writeFieldCount(count uint32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, count)
}

// private

// _grow grows the buffer and returns an element buffer of size `n`.
func (b *writeBuffer) _grow(n int) []byte {
	buf := (*b)

	// realloc
	rem := cap(buf) - len(buf)
	if rem < n {
		size := (cap(buf) * 2) + n
		next := make([]byte, len(buf), size)
		copy(next, buf)
		*b = next
		buf = (*b)
	}

	// grow buffer
	off := len(buf)
	ln := off + n
	*b = buf[:ln]

	// return element
	return buf[off:ln]
}
