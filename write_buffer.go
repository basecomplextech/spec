package protocol

import (
	"encoding/binary"
	"math"
)

type writeBuffer []byte

func (b *writeBuffer) offset() int {
	return len(*b)
}

// type

func (b *writeBuffer) type_(type_ Type) {
	p := b._grow(1)
	p[0] = byte(type_)
}

// primitives

func (b *writeBuffer) int8(v int8) {
	p := b._grow(1)
	p[0] = byte(v)
}

func (b *writeBuffer) int16(v int16) {
	p := b._grow(2)
	binary.BigEndian.PutUint16(p, uint16(v))
}

func (b *writeBuffer) int32(v int32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, uint32(v))
}

func (b *writeBuffer) int64(v int64) {
	p := b._grow(8)
	binary.BigEndian.PutUint64(p, uint64(v))
}

func (b *writeBuffer) uint8(v uint8) {
	p := b._grow(1)
	p[0] = v
}

func (b *writeBuffer) uint16(v uint16) {
	p := b._grow(2)
	binary.BigEndian.PutUint16(p, v)
}

func (b *writeBuffer) uint32(v uint32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, v)
}

func (b *writeBuffer) uint64(v uint64) {
	p := b._grow(8)
	binary.BigEndian.PutUint64(p, v)
}

func (b *writeBuffer) float32(v float32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, math.Float32bits(v))
}

func (b *writeBuffer) float64(v float64) {
	p := b._grow(8)
	binary.BigEndian.PutUint64(p, math.Float64bits(v))
}

// bytes

func (b *writeBuffer) bytes(v []byte) uint32 {
	size := len(v)
	p := b._grow(size)
	copy(p, v)
	return uint32(size)
}

func (b *writeBuffer) bytesSize(size uint32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, uint32(size))
}

// string

func (b *writeBuffer) string(s string) uint32 {
	size := len(s)
	p := b._grow(size)
	copy(p, s)
	return uint32(size)
}

func (b *writeBuffer) stringZero() {
	p := b._grow(1)
	p[0] = 0
}

func (b *writeBuffer) stringSize(size uint32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, uint32(size))
}

// list

func (b *writeBuffer) listTable(table []listElement) uint32 {
	size := len(table) * listElementSize
	p := b._grow(size)

	for i, elem := range table {
		off := i * listElementSize
		q := p[off : off+listElementSize]

		binary.BigEndian.PutUint32(q, elem.offset)
	}

	return uint32(size)
}

func (b *writeBuffer) listTableSize(size uint32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, size)
}

func (b *writeBuffer) listDataSize(size uint32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, size)
}

// message

func (b *writeBuffer) messageTable(table []messageField) uint32 {
	size := len(table) * messageFieldSize
	p := b._grow(size)

	for i, field := range table {
		off := i * messageFieldSize
		q := p[off : off+messageFieldSize]

		binary.BigEndian.PutUint16(q, field.tag)
		binary.BigEndian.PutUint32(q[2:], field.offset)
	}

	return uint32(size)
}

func (b *writeBuffer) messageTableSize(size uint32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, size)
}

func (b *writeBuffer) messageDataSize(size uint32) {
	p := b._grow(4)
	binary.BigEndian.PutUint32(p, size)
}

// private

// _grow grows the buffer and returns an element buffer of size `n`.
func (b *writeBuffer) _grow(n int) []byte {
	buf := (*b)

	// realloc
	rem := cap(buf) - len(buf)
	if rem < n {
		size := (cap(buf) * 2) + n

		p := make([]byte, len(buf), size)
		copy(p, buf)

		*b = p
		buf = (*b)
	}

	// grow buffer
	off := len(buf)
	ln := off + n
	*b = buf[:ln]

	// return element
	return buf[off:ln]
}
