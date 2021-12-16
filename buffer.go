package spec

import (
	"encoding/binary"
	"math"
	"unsafe"
)

type buffer []byte

// type and size

func (b buffer) peekType() Type {
	if len(b) == 0 {
		return TypeNil
	}

	off := len(b) - 1
	v := b[off]
	return Type(v)
}

func (b buffer) type_() (Type, buffer) {
	if len(b) == 0 {
		return TypeNil, nil
	}

	v := b.peekType()
	off := len(b) - 1
	return v, b[:off]
}

func (b buffer) size() (uint32, buffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

// primitives

func (b buffer) byte() (byte, buffer) {
	off := len(b) - 1
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := p[0]
	return v, b[:off]
}

func (b buffer) int8() (int8, buffer) {
	off := len(b) - 1
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := p[0]
	return int8(v), b[:off]
}

func (b buffer) int16() (int16, buffer) {
	off := len(b) - 2
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint16(p)
	return int16(v), b[:off]
}

func (b buffer) int32() (int32, buffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return int32(v), b[:off]
}

func (b buffer) int64() (int64, buffer) {
	off := len(b) - 8
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint64(p)
	return int64(v), b[:off]
}

func (b buffer) uint8() (uint8, buffer) {
	off := len(b) - 1
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := p[0]
	return v, b[:off]
}

func (b buffer) uint16() (uint16, buffer) {
	off := len(b) - 2
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint16(p)
	return v, b[:off]
}

func (b buffer) uint32() (uint32, buffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func (b buffer) uint64() (uint64, buffer) {
	off := len(b) - 8
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint64(p)
	return v, b[:off]
}

func (b buffer) float32() (float32, buffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return math.Float32frombits(v), b[:off]
}

func (b buffer) float64() (float64, buffer) {
	off := len(b) - 8
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint64(p)
	return math.Float64frombits(v), b[:off]
}

// bytes

func (b buffer) bytes(size uint32) ([]byte, buffer) {
	off := len(b) - int(size)
	if off < 0 {
		return nil, nil
	}

	p := b[off:]
	return p, b[:off]
}

// string

func (b buffer) string(size uint32) (string, buffer) {
	off := len(b) - int(size)
	if off < 0 {
		return "", nil
	}

	p := b[off:]
	s := *(*string)(unsafe.Pointer(&p))
	return s, b[:off]
}

// list

func (b buffer) listBytes(tableSize uint32, dataSize uint32) ([]byte, buffer) {
	size := int(1 + 4 + 4 + tableSize + dataSize) // type(1) + table size (4) + data size (4)
	off := len(b) - size
	if off < 0 {
		return nil, nil
	}

	return b[off:], b[:off]
}

func (b buffer) listTable(size uint32) (listTable, buffer) {
	off := len(b) - int(size)
	if off < 0 {
		return nil, nil
	}

	p := b[off:]
	v := listTable(p)
	return v, b[:off]
}

func (b buffer) listData(size uint32) (buffer, buffer) {
	off := len(b) - int(size)
	if off < 0 {
		return nil, nil
	}

	p := b[off:]
	return p, b[:off]
}

func (b buffer) listTableSize() (uint32, buffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func (b buffer) listDataSize() (uint32, buffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

// TODO: Replace with list data
func (b buffer) listElement(off uint32) []byte {
	if len(b) < int(off) {
		return nil
	}

	return b[:off]
}

// message

func (b buffer) messageBytes(tableSize uint32, dataSize uint32) ([]byte, buffer) {
	size := int(1 + 4 + 4 + tableSize + dataSize) // type(1) + table size(4) + data size(4)
	off := len(b) - size
	if off < 0 {
		return nil, nil
	}

	return b[off:], b[:off]
}

func (b buffer) messageTable(size uint32) (messageTable, buffer) {
	off := len(b) - int(size)
	if off < 0 {
		return nil, nil
	}

	p := b[off:]
	v := messageTable(p)
	return v, b[:off]
}

func (b buffer) messageData(size uint32) (messageData, buffer) {
	off := len(b) - int(size)
	if off < 0 {
		return messageData{}, nil
	}

	p := b[off:]
	return messageData{p}, b[:off]
}

func (b buffer) messageTableSize() (uint32, buffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func (b buffer) messageDataSize() (uint32, buffer) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}
