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
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

// primitives

func (b readBuffer) byte() (byte, readBuffer) {
	off := len(b) - 1
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := p[0]
	return v, b[:off]
}

func (b readBuffer) int8() (int8, readBuffer) {
	off := len(b) - 1
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := p[0]
	return int8(v), b[:off]
}

func (b readBuffer) int16() (int16, readBuffer) {
	off := len(b) - 2
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint16(p)
	return int16(v), b[:off]
}

func (b readBuffer) int32() (int32, readBuffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return int32(v), b[:off]
}

func (b readBuffer) int64() (int64, readBuffer) {
	off := len(b) - 8
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint64(p)
	return int64(v), b[:off]
}

func (b readBuffer) uint8() (uint8, readBuffer) {
	off := len(b) - 1
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := p[0]
	return v, b[:off]
}

func (b readBuffer) uint16() (uint16, readBuffer) {
	off := len(b) - 2
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint16(p)
	return v, b[:off]
}

func (b readBuffer) uint32() (uint32, readBuffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func (b readBuffer) uint64() (uint64, readBuffer) {
	off := len(b) - 8
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint64(p)
	return v, b[:off]
}

func (b readBuffer) float32() (float32, readBuffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return math.Float32frombits(v), b[:off]
}

func (b readBuffer) float64() (float64, readBuffer) {
	off := len(b) - 8
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint64(p)
	return math.Float64frombits(v), b[:off]
}

// bytes

func (b readBuffer) bytes(size uint32) ([]byte, readBuffer) {
	off := len(b) - int(size)
	if off < 0 {
		return nil, nil
	}

	p := b[off:]
	return p, b[:off]
}

// string

func (b readBuffer) string(size uint32) (string, readBuffer) {
	off := len(b) - int(size)
	if off < 0 {
		return "", nil
	}

	p := b[off:]
	s := *(*string)(unsafe.Pointer(&p))
	return s, b[:off]
}

// list

func (b readBuffer) listBytes(tableSize uint32, dataSize uint32) ([]byte, readBuffer) {
	size := int(1 + 4 + 4 + tableSize + dataSize) // type(1) + table size (4) + data size (4)
	off := len(b) - size
	if off < 0 {
		return nil, nil
	}

	return b[off:], b[:off]
}

func (b readBuffer) listTable(size uint32) (elementTable, readBuffer) {
	off := len(b) - int(size)
	if off < 0 {
		return nil, nil
	}

	p := b[off:]
	v := elementTable(p)
	return v, b[:off]
}

func (b readBuffer) listData(size uint32) (readBuffer, readBuffer) {
	off := len(b) - int(size)
	if off < 0 {
		return nil, nil
	}

	p := b[off:]
	return p, b[:off]
}

func (b readBuffer) listTableSize() (uint32, readBuffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func (b readBuffer) listDataSize() (uint32, readBuffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func (b readBuffer) listElement(off uint32) readBuffer {
	if len(b) < int(off) {
		return nil
	}

	return b[off:]
}

// message

func (b readBuffer) messageBytes(tableSize uint32, dataSize uint32) ([]byte, readBuffer) {
	size := int(1 + 4 + 4 + tableSize + dataSize) // type(1) + table size(4) + data size(4)
	off := len(b) - size
	if off < 0 {
		return nil, nil
	}

	return b[off:], b[:off]
}

func (b readBuffer) messageTable(size uint32) (fieldTable, readBuffer) {
	off := len(b) - int(size)
	if off < 0 {
		return nil, nil
	}

	p := b[off:]
	v := fieldTable(p)
	return v, b[:off]
}

func (b readBuffer) messageData(size uint32) (readBuffer, readBuffer) {
	off := len(b) - int(size)
	if off < 0 {
		return nil, nil
	}

	p := b[off:]
	return p, b[:off]
}

func (b readBuffer) messageTableSize() (uint32, readBuffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func (b readBuffer) messageDataSize() (uint32, readBuffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func (b readBuffer) messageField(off uint32) readBuffer {
	if len(b) < int(off) {
		return nil
	}
	return b[off:]
}
