package spec

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/baseblck/library/buffer"
	"github.com/baseblck/library/encoding/compactint"
	"github.com/baseblck/library/u128"
	"github.com/baseblck/library/u256"
)

type EncodeFunc[T any] func(b buffer.Buffer, value T) (int, error)

func EncodeNil(b buffer.Buffer) (int, error) {
	p := b.Grow(1)
	p[0] = byte(TypeNil)
	return 1, nil
}

func EncodeBool(b buffer.Buffer, v bool) (int, error) {
	p := b.Grow(1)
	if v {
		p[0] = byte(TypeTrue)
	} else {
		p[0] = byte(TypeFalse)
	}
	return 1, nil
}

func EncodeByte(b buffer.Buffer, v byte) (int, error) {
	p := b.Grow(2)
	p[0] = v
	p[1] = byte(TypeByte)
	return 2, nil
}

// Int

func EncodeInt32(b buffer.Buffer, v int32) (int, error) {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseInt32(p[:], v)
	off := compactint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(TypeInt32)

	return n + 1, nil
}

func EncodeInt64(b buffer.Buffer, v int64) (int, error) {
	p := [compactint.MaxLen64]byte{}
	n := compactint.PutReverseInt64(p[:], v)
	off := compactint.MaxLen64 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(TypeInt64)

	return n + 1, nil
}

// Uint

func EncodeUint32(b buffer.Buffer, v uint32) (int, error) {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], v)
	off := compactint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(TypeUint32)

	return n + 1, nil
}

func EncodeUint64(b buffer.Buffer, v uint64) (int, error) {
	p := [compactint.MaxLen64]byte{}
	n := compactint.PutReverseUint64(p[:], v)
	off := compactint.MaxLen64 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(TypeUint64)

	return n + 1, nil
}

// U128/U256

func EncodeU128(b buffer.Buffer, v u128.U128) (int, error) {
	p := b.Grow(17)
	copy(p, v[:])
	p[16] = byte(TypeU128)
	return 17, nil
}

func EncodeU256(b buffer.Buffer, v u256.U256) (int, error) {
	p := b.Grow(33)
	copy(p, v[:])
	p[32] = byte(TypeU256)
	return 33, nil
}

// Float

func EncodeFloat32(b buffer.Buffer, v float32) (int, error) {
	p := b.Grow(5)
	binary.BigEndian.PutUint32(p, math.Float32bits(v))
	p[4] = byte(TypeFloat32)
	return 5, nil
}

func EncodeFloat64(b buffer.Buffer, v float64) (int, error) {
	p := b.Grow(9)
	binary.BigEndian.PutUint64(p, math.Float64bits(v))
	p[8] = byte(TypeFloat64)
	return 9, nil
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

	big := size >= math.MaxUint16
	type_ := TypeBytes
	if big {
		type_ = TypeBytesBig
	}

	n += encodeSizeType(b, big, uint32(size), type_)
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

	big := size >= math.MaxUint16
	type_ := TypeString
	if big {
		type_ = TypeBigString
	}

	n += encodeSizeType(b, big, uint32(size), type_)
	return n, nil
}

// Struct

func EncodeStruct(b buffer.Buffer, dataSize int) (int, error) {
	if dataSize > MaxSize {
		return 0, fmt.Errorf("encode: struct too large, max size=%d, actual size=%d", MaxSize, dataSize)
	}

	big := dataSize >= math.MaxUint16
	type_ := TypeStruct
	if big {
		type_ = TypeBigStruct
	}

	n := encodeSizeType(b, big, uint32(dataSize), type_)
	return n, nil
}

// list meta

func encodeListMeta(b buffer.Buffer, dataSize int, table []listElement) (int, error) {
	if dataSize > MaxSize {
		return 0, fmt.Errorf("encode: list too large, max size=%d, actual size=%d", MaxSize, dataSize)
	}

	// type
	big := isBigList(table)
	type_ := TypeList
	if big {
		type_ = TypeBigList
	}

	// write table
	tableSize, err := encodeListTable(b, table, big)
	if err != nil {
		return int(tableSize), err
	}
	n := tableSize

	// write data size
	n += encodeSize(b, big, uint32(dataSize))

	// write table size and type
	n += encodeSizeType(b, big, uint32(tableSize), type_)
	return n, nil
}

func encodeListTable(b buffer.Buffer, table []listElement, big bool) (int, error) {
	// element size
	elemSize := listElementSmallSize
	if big {
		elemSize = listElementBigSize
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

	// type
	big := isBigMessage(table)
	type_ := TypeMessage
	if big {
		type_ = TypeBigMessage
	}

	// write table
	tableSize, err := encodeMessageTable(b, table, big)
	if err != nil {
		return 0, err
	}
	n := tableSize

	// write data size
	n += encodeSize(b, big, uint32(dataSize))

	// write table size and type
	n += encodeSizeType(b, big, uint32(tableSize), type_)
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

// private

// appendSize appends size as rvarint, for tests.
func appendSize(b []byte, big bool, size uint32) []byte {
	if big {
		p := [4]byte{}
		binary.BigEndian.PutUint32(p[:], size)
		return append(b, p[:]...)
	}

	if size >= math.MaxUint16 {
		panic("size too big")
	}

	p := [2]byte{}
	binary.BigEndian.PutUint16(p[:], uint16(size))
	return append(b, p[:]...)
}

func encodeSize(b buffer.Buffer, big bool, size uint32) int {
	if big {
		buf := b.Grow(4)
		binary.BigEndian.PutUint32(buf, size)
		return 4
	}

	if size >= math.MaxUint16 {
		panic("size too big")
	}

	buf := b.Grow(2)
	binary.BigEndian.PutUint16(buf, uint16(size))
	return 2
}

func encodeSizeType(b buffer.Buffer, big bool, size uint32, type_ Type) int {
	if big {
		buf := b.Grow(5)
		binary.BigEndian.PutUint32(buf, size)
		buf[4] = byte(type_)
		return 5
	}

	if size >= math.MaxUint16 {
		panic("size too big")
	}

	buf := b.Grow(3)
	binary.BigEndian.PutUint16(buf, uint16(size))
	buf[2] = byte(type_)
	return 3
}
