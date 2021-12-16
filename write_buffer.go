package protocol

import (
	"encoding/binary"
	"math"
)

type writeBuffer struct {
	buffer []byte

	_scratch [16]byte
}

func (b *writeBuffer) reset() {
	b.buffer = nil
}

func (b *writeBuffer) offset() int {
	return len(b.buffer)
}

// type

func (b *writeBuffer) type_(type_ Type) {
	p := b._scratch[:1]
	p[0] = byte(type_)

	b.buffer = append(b.buffer, p...)
}

// primitives

func (b *writeBuffer) int8(v int8) {
	p := b._scratch[:1]
	p[0] = byte(v)

	b.buffer = append(b.buffer, p...)
}

func (b *writeBuffer) int16(v int16) {
	p := b._scratch[:2]
	binary.BigEndian.PutUint16(p, uint16(v))

	b.buffer = append(b.buffer, p...)
}

func (b *writeBuffer) int32(v int32) {
	p := b._scratch[:4]
	binary.BigEndian.PutUint32(p, uint32(v))

	b.buffer = append(b.buffer, p...)
}

func (b *writeBuffer) int64(v int64) {
	p := b._scratch[:8]
	binary.BigEndian.PutUint64(p, uint64(v))

	b.buffer = append(b.buffer, p...)
}

func (b *writeBuffer) uint8(v uint8) {
	p := b._scratch[:1]
	p[0] = v

	b.buffer = append(b.buffer, p...)
}

func (b *writeBuffer) uint16(v uint16) {
	p := b._scratch[:2]
	binary.BigEndian.PutUint16(p, v)

	b.buffer = append(b.buffer, p...)
}

func (b *writeBuffer) uint32(v uint32) {
	p := b._scratch[:4]
	binary.BigEndian.PutUint32(p, v)

	b.buffer = append(b.buffer, p...)
}

func (b *writeBuffer) uint64(v uint64) {
	p := b._scratch[:8]
	binary.BigEndian.PutUint64(p, v)

	b.buffer = append(b.buffer, p...)
}

func (b *writeBuffer) float32(v float32) {
	p := b._scratch[:4]
	binary.BigEndian.PutUint32(p, math.Float32bits(v))

	b.buffer = append(b.buffer, p...)
}

func (b *writeBuffer) float64(v float64) {
	p := b._scratch[:8]
	binary.BigEndian.PutUint64(p, math.Float64bits(v))

	b.buffer = append(b.buffer, p...)
}

// bytes

func (b *writeBuffer) bytes(v []byte) uint32 {
	b.buffer = append(b.buffer, v...)
	return uint32(len(v))
}

func (b *writeBuffer) bytesSize(size uint32) {
	p := b._scratch[:4]
	binary.BigEndian.PutUint32(p, uint32(size))

	b.buffer = append(b.buffer, p...)
}

// string

func (b *writeBuffer) string(s string) uint32 {
	size := uint32(len(s))
	b.buffer = append(b.buffer, s...)
	return uint32(size)
}

func (b *writeBuffer) stringZero() {
	p := b._scratch[:1]
	p[0] = 0

	b.buffer = append(b.buffer, p...)
}

func (b *writeBuffer) stringSize(size uint32) {
	p := b._scratch[:4]
	binary.BigEndian.PutUint32(p, uint32(size))

	b.buffer = append(b.buffer, p...)
}

// list

func (b *writeBuffer) listTable(table []listElement) uint32 {
	size := len(table) * listElementSize

	for _, elem := range table {
		p := b._scratch[:listElementSize]
		binary.BigEndian.PutUint32(p, elem.offset)

		b.buffer = append(b.buffer, p...)
	}

	return uint32(size)
}

func (b *writeBuffer) listTableSize(size uint32) {
	p := b._scratch[:4]
	binary.BigEndian.PutUint32(p, size)

	b.buffer = append(b.buffer, p...)
}

func (b *writeBuffer) listDataSize(size uint32) {
	p := b._scratch[:4]
	binary.BigEndian.PutUint32(p, size)

	b.buffer = append(b.buffer, p...)
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
	p := b._scratch[:4]
	binary.BigEndian.PutUint32(p, size)

	b.buffer = append(b.buffer, p...)
}

func (b *writeBuffer) messageDataSize(size uint32) {
	p := b._scratch[:4]
	binary.BigEndian.PutUint32(p, size)

	b.buffer = append(b.buffer, p...)
}

// private

func (b *writeBuffer) _grow(n int) []byte {
	cp := cap(b.buffer)
	ln := len(b.buffer)

	// alloc
	rem := cp - ln
	if rem < n {
		size := (cp * 2) + n
		buf := make([]byte, ln, size)
		copy(buf, b.buffer)
		b.buffer = buf
	}

	// return
	size := ln + n
	b.buffer = b.buffer[:size]
	p := b.buffer[ln:size]
	return p
}
