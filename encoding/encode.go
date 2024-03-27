package encoding

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/baselibrary/encoding/compactint"
	"github.com/basecomplextech/spec/internal/core"
)

func EncodeBool(b buffer.Buffer, v bool) (int, error) {
	p := b.Grow(1)
	if v {
		p[0] = byte(core.TypeTrue)
	} else {
		p[0] = byte(core.TypeFalse)
	}
	return 1, nil
}

func EncodeByte(b buffer.Buffer, v byte) (int, error) {
	p := b.Grow(2)
	p[0] = v
	p[1] = byte(core.TypeByte)
	return 2, nil
}

// Int

func EncodeInt16(b buffer.Buffer, v int16) (int, error) {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseInt32(p[:], int32(v))
	off := compactint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(core.TypeInt16)

	return n + 1, nil
}

func EncodeInt32(b buffer.Buffer, v int32) (int, error) {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseInt32(p[:], v)
	off := compactint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(core.TypeInt32)

	return n + 1, nil
}

func EncodeInt64(b buffer.Buffer, v int64) (int, error) {
	p := [compactint.MaxLen64]byte{}
	n := compactint.PutReverseInt64(p[:], v)
	off := compactint.MaxLen64 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(core.TypeInt64)

	return n + 1, nil
}

// Uint

func EncodeUint16(b buffer.Buffer, v uint16) (int, error) {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], uint32(v))
	off := compactint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(core.TypeUint16)

	return n + 1, nil
}

func EncodeUint32(b buffer.Buffer, v uint32) (int, error) {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], v)
	off := compactint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(core.TypeUint32)

	return n + 1, nil
}

func EncodeUint64(b buffer.Buffer, v uint64) (int, error) {
	p := [compactint.MaxLen64]byte{}
	n := compactint.PutReverseUint64(p[:], v)
	off := compactint.MaxLen64 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(core.TypeUint64)

	return n + 1, nil
}

// Bin64/128/256

func EncodeBin64(b buffer.Buffer, v bin.Bin64) (int, error) {
	p := b.Grow(9)
	copy(p, v[:])
	p[8] = byte(core.TypeBin64)
	return 9, nil
}

func EncodeBin128(b buffer.Buffer, v bin.Bin128) (int, error) {
	p := b.Grow(17)
	copy(p, v[:])
	p[16] = byte(core.TypeBin128)
	return 17, nil
}

func EncodeBin256(b buffer.Buffer, v bin.Bin256) (int, error) {
	p := b.Grow(33)
	copy(p, v[:])
	p[32] = byte(core.TypeBin256)
	return 33, nil
}

// Float

func EncodeFloat32(b buffer.Buffer, v float32) (int, error) {
	p := b.Grow(5)
	binary.BigEndian.PutUint32(p, math.Float32bits(v))
	p[4] = byte(core.TypeFloat32)
	return 5, nil
}

func EncodeFloat64(b buffer.Buffer, v float64) (int, error) {
	p := b.Grow(9)
	binary.BigEndian.PutUint64(p, math.Float64bits(v))
	p[8] = byte(core.TypeFloat64)
	return 9, nil
}

// Bytes

func EncodeBytes(b buffer.Buffer, v []byte) (int, error) {
	size := len(v)
	if size > core.MaxSize {
		return 0, fmt.Errorf("encode: bytes too large, max size=%d, actual size=%d", core.MaxSize, size)
	}

	p := b.Grow(size)
	copy(p, v)
	n := size

	n += encodeSizeType(b, uint32(size), core.TypeBytes)
	return n, nil
}

// String

func EncodeString(b buffer.Buffer, s string) (int, error) {
	size := len(s)
	if size > core.MaxSize {
		return 0, fmt.Errorf("encode: string too large, max size=%d, actual size=%d", core.MaxSize, size)
	}

	n := size + 1 // plus zero byte
	p := b.Grow(n)
	copy(p, s)

	n += encodeSizeType(b, uint32(size), core.TypeString)
	return n, nil
}

// Struct

func EncodeStruct(b buffer.Buffer, dataSize int) (int, error) {
	if dataSize > core.MaxSize {
		return 0, fmt.Errorf("encode: struct too large, max size=%d, actual size=%d", core.MaxSize, dataSize)
	}

	n := encodeSizeType(b, uint32(dataSize), core.TypeStruct)
	return n, nil
}

// ListMeta

func EncodeListMeta(b buffer.Buffer, dataSize int, table []ListElement) (int, error) {
	if dataSize > core.MaxSize {
		return 0, fmt.Errorf("encode: list too large, max size=%d, actual size=%d", core.MaxSize, dataSize)
	}

	// core.Type
	big := isBigList(table)
	type_ := core.TypeList
	if big {
		type_ = core.TypeBigList
	}

	// Write table
	tableSize, err := encodeListTable(b, table, big)
	if err != nil {
		return int(tableSize), err
	}
	n := tableSize

	// Write data size
	n += encodeSize(b, uint32(dataSize))

	// Write table size and type
	n += encodeSizeType(b, uint32(tableSize), type_)
	return n, nil
}

func encodeListTable(b buffer.Buffer, table []ListElement, big bool) (int, error) {
	// Element size
	elemSize := listElementSmallSize
	if big {
		elemSize = listElementBigSize
	}

	// Check table size
	size := len(table) * elemSize
	if size > core.MaxSize {
		return 0, fmt.Errorf("encode: list table too large, max size=%d, actual size=%d", core.MaxSize, size)
	}

	// Write table
	p := b.Grow(size)
	off := 0

	// Put elements
	for _, elem := range table {
		q := p[off : off+elemSize]

		if big {
			binary.BigEndian.PutUint32(q, elem.Offset)
		} else {
			binary.BigEndian.PutUint16(q, uint16(elem.Offset))
		}

		off += elemSize
	}

	return size, nil
}

// MessageMeta

func EncodeMessageMeta(b buffer.Buffer, dataSize int, table []MessageField) (int, error) {
	if dataSize > core.MaxSize {
		return 0, fmt.Errorf("encode: message too large, max size=%d, actual size=%d", core.MaxSize, dataSize)
	}

	// core.Type
	big := isBigMessage(table)
	type_ := core.TypeMessage
	if big {
		type_ = core.TypeBigMessage
	}

	// Write table
	tableSize, err := encodeMessageTable(b, table, big)
	if err != nil {
		return 0, err
	}
	n := tableSize

	// Write data size
	n += encodeSize(b, uint32(dataSize))

	// Write table size and type
	n += encodeSizeType(b, uint32(tableSize), type_)
	return n, nil
}

func encodeMessageTable(b buffer.Buffer, table []MessageField, big bool) (int, error) {
	// Field size
	var fieldSize int
	if big {
		fieldSize = messageFieldBigSize
	} else {
		fieldSize = messageFieldSmallSize
	}

	// Check table size
	size := len(table) * fieldSize
	if size > core.MaxSize {
		return 0, fmt.Errorf("encode: message table too large, max size=%d, actual size=%d", core.MaxSize, size)
	}

	// Write table
	p := b.Grow(size)
	off := 0

	// Put fields
	for _, field := range table {
		q := p[off : off+fieldSize]

		if big {
			binary.BigEndian.PutUint16(q, field.Tag)
			binary.BigEndian.PutUint32(q[2:], field.Offset)
		} else {
			q[0] = byte(field.Tag)
			binary.BigEndian.PutUint16(q[1:], uint16(field.Offset))
		}

		off += fieldSize
	}

	return size, nil
}

// private

// appendSize appends size as rvarint, for tests.
func appendSize(b []byte, big bool, size uint32) []byte {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], size)
	off := compactint.MaxLen32 - n

	return append(b, p[off:]...)
}

func encodeSize(b buffer.Buffer, size uint32) int {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], size)
	off := compactint.MaxLen32 - n

	buf := b.Grow(n)
	copy(buf, p[off:])

	return n
}

func encodeSizeType(b buffer.Buffer, size uint32, type_ core.Type) int {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], size)
	off := compactint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(type_)

	return n + 1
}
