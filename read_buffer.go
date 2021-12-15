package protocol

import (
	"encoding/binary"
	"math"
	"unsafe"
)

type readBuffer []byte

// type and size

func (b readBuffer) peekType() Type {
	if len(b) == 0 {
		return TypeNil
	}

	off := len(b) - 1
	v := b[off]
	return Type(v)
}

func (b readBuffer) type_() (Type, readBuffer) {
	if len(b) == 0 {
		return TypeNil, nil
	}

	v := b.peekType()
	off := len(b) - 1
	return v, b[:off]
}

func (b readBuffer) size() (uint32, readBuffer) {
	if len(b) < 4 {
		return 0, nil
	}

	off := len(b) - 4
	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

// primitives

func (b readBuffer) byte() (byte, readBuffer) {
	if len(b) == 0 {
		return 0, nil
	}

	off := len(b) - 1
	p := b[off:]
	v := p[0]
	return v, b[:off]
}

func (b readBuffer) int8() (int8, readBuffer) {
	if len(b) == 0 {
		return 0, nil
	}

	off := len(b) - 1
	p := b[off:]
	v := p[0]
	return int8(v), b[:off]
}

func (b readBuffer) int16() (int16, readBuffer) {
	if len(b) < 2 {
		return 0, nil
	}

	off := len(b) - 2
	p := b[off:]
	v := binary.BigEndian.Uint16(p)
	return int16(v), b[:off]
}

func (b readBuffer) int32() (int32, readBuffer) {
	if len(b) < 4 {
		return 0, nil
	}

	off := len(b) - 4
	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return int32(v), b[:off]
}

func (b readBuffer) int64() (int64, readBuffer) {
	if len(b) < 8 {
		return 0, nil
	}

	off := len(b) - 8
	p := b[off:]
	v := binary.BigEndian.Uint64(p)
	return int64(v), b[:off]
}

func (b readBuffer) uint8() (uint8, readBuffer) {
	if len(b) == 0 {
		return 0, nil
	}

	off := len(b) - 1
	p := b[off:]
	v := p[0]
	return v, b[:off]
}

func (b readBuffer) uint16() (uint16, readBuffer) {
	if len(b) < 2 {
		return 0, nil
	}

	off := len(b) - 2
	p := b[off:]
	v := binary.BigEndian.Uint16(p)
	return v, b[:off]
}

func (b readBuffer) uint32() (uint32, readBuffer) {
	if len(b) < 4 {
		return 0, nil
	}

	off := len(b) - 4
	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func (b readBuffer) uint64() (uint64, readBuffer) {
	if len(b) < 8 {
		return 0, nil
	}

	off := len(b) - 8
	p := b[off:]
	v := binary.BigEndian.Uint64(p)
	return v, b[:off]
}

func (b readBuffer) float32() (float32, readBuffer) {
	if len(b) < 4 {
		return 0, nil
	}

	off := len(b) - 4
	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return math.Float32frombits(v), b[:off]
}

func (b readBuffer) float64() (float64, readBuffer) {
	if len(b) < 8 {
		return 0, nil
	}

	off := len(b) - 8
	p := b[off:]
	v := binary.BigEndian.Uint64(p)
	return math.Float64frombits(v), b[:off]
}

// bytes

func (b readBuffer) bytes(size uint32) ([]byte, readBuffer) {
	if len(b) < int(size) {
		return nil, nil
	}

	off := len(b) - int(size)
	p := b[off:]
	return p, b[:off]
}

// string

func (b readBuffer) string(size uint32) (string, readBuffer) {
	if len(b) < int(size) {
		return "", nil
	}

	off := len(b) - int(size)
	p := b[off:]
	s := *(*string)(unsafe.Pointer(&p))
	return s, b[:off]
}

// list

func (b readBuffer) listSize() (uint32, readBuffer) {
	if len(b) < 4 {
		return 0, nil
	}

	off := len(b) - 4
	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func (b readBuffer) listCount() (uint32, readBuffer) {
	if len(b) < 4 {
		return 0, nil
	}

	off := len(b) - 4
	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func (b readBuffer) listTable(count uint32) (elementTable, readBuffer) {
	size := int(count * elementSize)
	if len(b) < size {
		return nil, nil
	}

	off := len(b) - size
	p := b[off:]
	v := elementTable(p)
	return v, b[:off]
}

func (b readBuffer) listData(size uint32, count uint32) (readBuffer, readBuffer) {
	ln := int(size - 4 - (count * elementSize))
	if len(b) < ln {
		return nil, nil
	}

	off := len(b) - ln
	p := b[off:]
	return p, b[:off]
}

func (b readBuffer) listElement(off uint32) readBuffer {
	if len(b) < int(off) {
		return nil
	}

	return b[off:]
}

// message

func (b readBuffer) messageSize() (uint32, readBuffer) {
	if len(b) < 4 {
		return 0, nil
	}

	off := len(b) - 4
	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func (b readBuffer) messageCount() (uint32, readBuffer) {
	if len(b) < 4 {
		return 0, nil
	}

	off := len(b) - 4
	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func (b readBuffer) messageTable(count uint32) (fieldTable, readBuffer) {
	size := int(count * fieldSize)
	if len(b) < size {
		return nil, nil
	}

	off := len(b) - size
	p := b[off:]
	v := fieldTable(p)
	return v, b[:off]
}

func (b readBuffer) messageData(size uint32, count uint32) (readBuffer, readBuffer) {
	ln := int(size - 4 - (count * fieldSize))
	if len(b) < ln {
		return nil, nil
	}

	off := len(b) - ln
	p := b[off:]
	return p, b[:off]
}

func (b readBuffer) messageField(off uint32) readBuffer {
	if len(b) < int(off) {
		return nil
	}
	return b[off:]
}
