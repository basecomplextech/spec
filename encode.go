package spec

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/complexl/library/buffer"
	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
	"github.com/complexl/spec/rvarint"
)

func EncodeNil(b buffer.Buffer) int {
	p := b.Grow(1)
	p[0] = byte(TypeNil)
	return 1
}

func EncodeBool(b buffer.Buffer, v bool) int {
	p := b.Grow(1)
	if v {
		p[0] = byte(TypeTrue)
	} else {
		p[0] = byte(TypeFalse)
	}
	return 1
}

func EncodeByte(b buffer.Buffer, v byte) int {
	p := b.Grow(2)
	p[0] = v
	p[1] = byte(TypeByte)
	return 2
}

// Int

func EncodeInt32(b buffer.Buffer, v int32) int {
	p := [rvarint.MaxLen32]byte{}
	n := rvarint.PutInt64(p[:], int64(v))
	off := rvarint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(TypeInt32)

	return n + 1
}

func EncodeInt64(b buffer.Buffer, v int64) int {
	p := [rvarint.MaxLen64]byte{}
	n := rvarint.PutInt64(p[:], v)
	off := rvarint.MaxLen64 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(TypeInt64)

	return n + 1
}

// Uint

func EncodeUint32(b buffer.Buffer, v uint32) int {
	p := [rvarint.MaxLen32]byte{}
	n := rvarint.PutUint64(p[:], uint64(v))
	off := rvarint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(TypeUint32)

	return n + 1
}

func EncodeUint64(b buffer.Buffer, v uint64) int {
	p := [rvarint.MaxLen64]byte{}
	n := rvarint.PutUint64(p[:], v)
	off := rvarint.MaxLen64 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(TypeUint64)

	return n + 1
}

// U128/U256

func EncodeU128(b buffer.Buffer, v u128.U128) int {
	p := b.Grow(17)
	copy(p, v[:])
	p[16] = byte(TypeU128)
	return 17
}

func EncodeU256(b buffer.Buffer, v u256.U256) int {
	p := b.Grow(33)
	copy(p, v[:])
	p[32] = byte(TypeU256)
	return 33
}

// Float

func EncodeFloat32(b buffer.Buffer, v float32) int {
	p := b.Grow(5)
	binary.BigEndian.PutUint32(p, math.Float32bits(v))
	p[4] = byte(TypeFloat32)
	return 5
}

func EncodeFloat64(b buffer.Buffer, v float64) int {
	p := b.Grow(9)
	binary.BigEndian.PutUint64(p, math.Float64bits(v))
	p[8] = byte(TypeFloat64)
	return 9
}

// Bytes

func EncodeBytes(b buffer.Buffer, v []byte) (int, error) {
	size := len(v)
	if size > MaxSize {
		return 0, fmt.Errorf("encode: bytes too large, max size=%d, actual size=%d", MaxSize, size)
	}

	p := b.Grow(size)
	copy(p, v)
	n := size

	n += encodeSizeType(b, uint32(size), TypeBytes)
	return n, nil
}

// String

func EncodeString(b buffer.Buffer, s string) (int, error) {
	size := len(s)
	if size > MaxSize {
		return 0, fmt.Errorf("encode: string too large, max size=%d, actual size=%d", MaxSize, size)
	}

	n := size + 1 // plus zero byte
	p := b.Grow(n)
	copy(p, s)

	n += encodeSizeType(b, uint32(size), TypeString)
	return n, nil
}

// list meta

func encodeListMeta(b buffer.Buffer, dataSize int, table []listElement) (int, error) {
	if dataSize > MaxSize {
		return 0, fmt.Errorf("encode: list too large, max size=%d, actual size=%d", MaxSize, dataSize)
	}

	// get type
	big := isBigList(table)
	var type_ Type
	if big {
		type_ = TypeListBig
	} else {
		type_ = TypeList
	}

	// write table
	tableSize, err := encodeListTable(b, table, big)
	if err != nil {
		return int(tableSize), err
	}
	n := tableSize

	// write data size
	n += encodeSize(b, uint32(dataSize))

	// write table size and type
	n += encodeSizeType(b, uint32(tableSize), type_)
	return n, nil
}

func encodeListTable(b buffer.Buffer, table []listElement, big bool) (int, error) {
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
		return 0, fmt.Errorf("encode: list table too large, max size=%d, actual size=%d", MaxSize, size)
	}

	// write table
	p := b.Grow(size)
	off := 0

	// put elements
	for _, elem := range table {
		q := p[off : off+elemSize]

		if big {
			binary.BigEndian.PutUint32(q, elem.offset)
		} else {
			binary.BigEndian.PutUint16(q, uint16(elem.offset))
		}

		off += elemSize
	}

	return size, nil
}

// message meta

func encodeMessageMeta(b buffer.Buffer, dataSize int, table []messageField) (int, error) {
	if dataSize > MaxSize {
		return 0, fmt.Errorf("encode: message too large, max size=%d, actual size=%d", MaxSize, dataSize)
	}

	// get type
	big := isBigMessage(table)
	var type_ Type
	if big {
		type_ = TypeMessageBig
	} else {
		type_ = TypeMessage
	}

	// write table
	tableSize, err := encodeMessageTable(b, table, big)
	if err != nil {
		return 0, err
	}
	n := tableSize

	// write data size
	n += encodeSize(b, uint32(dataSize))

	// write table size and type
	n += encodeSizeType(b, uint32(tableSize), type_)
	return n, nil
}

func encodeMessageTable(b buffer.Buffer, table []messageField, big bool) (int, error) {
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
		return 0, fmt.Errorf("encode: message table too large, max size=%d, actual size=%d", MaxSize, size)
	}

	// write table
	p := b.Grow(size)
	off := 0

	// put fields
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

	return size, nil
}

// struct

func encodeStruct(b buffer.Buffer, dataSize int) (int, error) {
	if dataSize > MaxSize {
		return 0, fmt.Errorf("encode: struct too large, max size=%d, actual size=%d", MaxSize, dataSize)
	}

	// write size and type
	n := encodeSizeType(b, uint32(dataSize), TypeStruct)
	return n, nil
}

// private

// appendSize appends size as rvarint, for tests.
func appendSize(b []byte, size uint32) []byte {
	p := [rvarint.MaxLen32]byte{}
	n := rvarint.PutUint64(p[:], uint64(size))
	off := rvarint.MaxLen32 - n
	return append(b, p[off:]...)
}

func encodeSize(b buffer.Buffer, size uint32) int {
	p := [rvarint.MaxLen32]byte{}
	n := rvarint.PutUint64(p[:], uint64(size))
	off := rvarint.MaxLen32 - n

	buf := b.Grow(n)
	copy(buf, p[off:])
	return n
}

func encodeSizeType(b buffer.Buffer, size uint32, type_ Type) int {
	p := [rvarint.MaxLen32]byte{}
	n := rvarint.PutUint64(p[:], uint64(size))
	off := rvarint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(type_)

	return n + 1
}
