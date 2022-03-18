package spec

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
	"github.com/complexl/spec/rvarint"
)

func EncodeNil(b []byte) []byte {
	return append(b, byte(TypeNil))
}

func EncodeBool(b []byte, v bool) []byte {
	if v {
		return append(b, byte(TypeTrue))
	} else {
		return append(b, byte(TypeFalse))
	}
}

func EncodeByte(b []byte, v byte) []byte {
	b = append(b, v)
	b = append(b, byte(TypeByte))
	return b
}

func EncodeInt32(b []byte, v int32) []byte {
	p := [rvarint.MaxLen32]byte{}
	n := rvarint.PutInt64(p[:], int64(v))
	off := rvarint.MaxLen32 - n

	b = append(b, p[off:]...)
	b = append(b, byte(TypeInt32))
	return b
}

func EncodeInt64(b []byte, v int64) []byte {
	p := [rvarint.MaxLen64]byte{}
	n := rvarint.PutInt64(p[:], v)
	off := rvarint.MaxLen64 - n

	b = append(b, p[off:]...)
	b = append(b, byte(TypeInt64))
	return b
}

func EncodeUint32(b []byte, v uint32) []byte {
	p := [rvarint.MaxLen32]byte{}
	n := rvarint.PutUint64(p[:], uint64(v))
	off := rvarint.MaxLen32 - n

	b = append(b, p[off:]...)
	b = append(b, byte(TypeUint32))
	return b
}

func EncodeUint64(b []byte, v uint64) []byte {
	p := [rvarint.MaxLen64]byte{}
	n := rvarint.PutUint64(p[:], v)
	off := rvarint.MaxLen64 - n

	b = append(b, p[off:]...)
	b = append(b, byte(TypeUint64))
	return b
}

// U128/U256

func EncodeU128(b []byte, v u128.U128) []byte {
	b = append(b, v[:]...)
	b = append(b, byte(TypeU128))
	return b
}

func EncodeU256(b []byte, v u256.U256) []byte {
	b = append(b, v[:]...)
	b = append(b, byte(TypeU256))
	return b
}

// Float

func EncodeFloat32(b []byte, v float32) []byte {
	p := [4]byte{}
	binary.BigEndian.PutUint32(p[:], math.Float32bits(v))

	b = append(b, p[:]...)
	b = append(b, byte(TypeFloat32))
	return b
}

func EncodeFloat64(b []byte, v float64) []byte {
	p := [8]byte{}
	binary.BigEndian.PutUint64(p[:], math.Float64bits(v))

	b = append(b, p[:]...)
	b = append(b, byte(TypeFloat64))
	return b
}

// Bytes

func EncodeBytes(b []byte, v []byte) ([]byte, error) {
	size := len(v)
	if size > MaxSize {
		return nil, fmt.Errorf("write: bytes too large, max size=%d, actual size=%d", MaxSize, size)
	}

	b = append(b, v...)
	b = encodeBytesSize(b, uint32(size))
	b = append(b, byte(TypeBytes))
	return b, nil
}

func encodeBytesSize(b []byte, size uint32) []byte {
	p := [rvarint.MaxLen32]byte{}
	n := rvarint.PutUint64(p[:], uint64(size))
	off := rvarint.MaxLen32 - n
	return append(b, p[off:]...)
}

// String

func EncodeString(b []byte, s string) ([]byte, error) {
	size := len(s)
	if size > MaxSize {
		return nil, fmt.Errorf("write: string too large, max size=%d, actual size=%d", MaxSize, size)
	}

	b = append(b, s...)
	b = append(b, 0) // zero byte
	b = encodeStringSize(b, uint32(size))
	b = append(b, byte(TypeString))
	return b, nil
}

func encodeStringSize(b []byte, size uint32) []byte {
	p := [rvarint.MaxLen32]byte{}
	n := rvarint.PutUint64(p[:], uint64(size))
	off := rvarint.MaxLen32 - n
	return append(b, p[off:]...)
}

// List

func encodeList(b []byte, bodySize int, table []listElement) ([]byte, error) {
	if bodySize > MaxSize {
		return nil, fmt.Errorf("write: list too large, max size=%d, actual size=%d", MaxSize, bodySize)
	}

	// type
	big := isBigList(table)
	var type_ Type
	if big {
		type_ = TypeListBig
	} else {
		type_ = TypeList
	}

	// sizes
	bsize := uint32(bodySize)
	tsize := uint32(0)

	// write table
	var err error
	b, tsize, err = encodeListTable(b, table, big)
	if err != nil {
		return nil, err
	}

	// write sizes and type
	b = encodeListBodySize(b, bsize)
	b = encodeListTableSize(b, tsize)
	b = append(b, byte(type_))
	return b, nil
}

func encodeListTable(b []byte, table []listElement, big bool) ([]byte, uint32, error) {
	// element size
	var elemSize int
	if big {
		elemSize = listElementBigSize
	} else {
		elemSize = listElementSmallSize
	}

	// check table size
	size := len(table) * elemSize
	if size > MaxSize {
		return nil, 0, fmt.Errorf("write: list table too large, max size=%d, actual size=%d", MaxSize, size)
	}

	// alloc table
	b, p := encodeGrow(b, size)
	off := 0

	// write elements
	for _, elem := range table {
		q := p[off : off+elemSize]

		if big {
			binary.BigEndian.PutUint32(q, elem.offset)
		} else {
			binary.BigEndian.PutUint16(q, uint16(elem.offset))
		}

		off += elemSize
	}

	return b, uint32(size), nil
}

func encodeListTableSize(b []byte, size uint32) []byte {
	p := [rvarint.MaxLen32]byte{}
	n := rvarint.PutUint64(p[:], uint64(size))
	off := rvarint.MaxLen32 - n
	return append(b, p[off:]...)
}

func encodeListBodySize(b []byte, size uint32) []byte {
	p := [rvarint.MaxLen32]byte{}
	n := rvarint.PutUint64(p[:], uint64(size))
	off := rvarint.MaxLen32 - n
	return append(b, p[off:]...)
}

// Message

func encodeMessage(b []byte, bodySize int, table []messageField) ([]byte, error) {
	if bodySize > MaxSize {
		return nil, fmt.Errorf("write: message too large, max size=%d, actual size=%d", MaxSize, bodySize)
	}

	// type
	big := isBigMessage(table)
	var type_ Type
	if big {
		type_ = TypeMessageBig
	} else {
		type_ = TypeMessage
	}

	// sizes
	bsize := uint32(bodySize)
	tsize := uint32(0)

	// write table
	var err error
	b, tsize, err = encodeMessageTable(b, table, big)
	if err != nil {
		return nil, err
	}

	// write sizes and type
	b = encodeMessageBodySize(b, bsize)
	b = encodeMessageTableSize(b, tsize)
	b = append(b, byte(type_))
	return b, nil
}

func encodeMessageTable(b []byte, table []messageField, big bool) ([]byte, uint32, error) {
	// field size
	var fieldSize int
	if big {
		fieldSize = messageFieldBigSize
	} else {
		fieldSize = messageFieldSmallSize
	}

	// check table size
	size := len(table) * fieldSize
	if size > MaxSize {
		return nil, 0, fmt.Errorf("write: message table too large, max size=%d, actual size=%d", MaxSize, size)
	}

	// alloc table
	b, p := encodeGrow(b, size)
	off := 0

	// write fields
	for _, field := range table {
		q := p[off : off+fieldSize]

		if big {
			binary.BigEndian.PutUint16(q, field.tag)
			binary.BigEndian.PutUint32(q[2:], field.offset)
		} else {
			q[0] = byte(field.tag)
			binary.BigEndian.PutUint16(q[1:], uint16(field.offset))
		}

		off += fieldSize
	}

	return b, uint32(size), nil
}

func encodeMessageTableSize(b []byte, size uint32) []byte {
	p := [rvarint.MaxLen32]byte{}
	n := rvarint.PutUint64(p[:], uint64(size))
	off := rvarint.MaxLen32 - n
	return append(b, p[off:]...)
}

func encodeMessageBodySize(b []byte, size uint32) []byte {
	p := [rvarint.MaxLen32]byte{}
	n := rvarint.PutUint64(p[:], uint64(size))
	off := rvarint.MaxLen32 - n
	return append(b, p[off:]...)
}

// Struct

func encodeStruct(b []byte, bodySize int) ([]byte, error) {
	if bodySize > MaxSize {
		return nil, fmt.Errorf("write: struct too large, max size=%d, actual size=%d", MaxSize, bodySize)
	}

	bsize := uint32(bodySize)
	b = encodeStructBodySize(b, bsize)
	b = append(b, byte(TypeStruct))
	return b, nil
}

func encodeStructBodySize(b []byte, size uint32) []byte {
	p := [rvarint.MaxLen32]byte{}
	n := rvarint.PutUint64(p[:], uint64(size))
	off := rvarint.MaxLen32 - n
	return append(b, p[off:]...)
}

// private

// encodeGrow grows a buffer by n bytes and returns a new buffer and an allocated slice.
func encodeGrow(b []byte, n int) ([]byte, []byte) {
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
