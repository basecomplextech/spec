package spec

import (
	"encoding/binary"
	"fmt"
	"math"
)

func writeBool(b []byte, v bool) []byte {
	if v {
		return append(b, byte(TypeTrue))
	} else {
		return append(b, byte(TypeFalse))
	}
}

func writeInt8(b []byte, v int8) []byte {
	b = append(b, uint8(v))
	b = append(b, byte(TypeInt8))
	return b
}

func writeInt16(b []byte, v int16) []byte {
	p := [maxVarintLen16]byte{}
	n := writeReverseVarint(p[:], int64(v))
	off := maxVarintLen16 - n

	b = append(b, p[off:]...)
	b = append(b, byte(TypeInt16))
	return b
}

func writeInt32(b []byte, v int32) []byte {
	p := [maxVarintLen32]byte{}
	n := writeReverseVarint(p[:], int64(v))
	off := maxVarintLen32 - n

	b = append(b, p[off:]...)
	b = append(b, byte(TypeInt32))
	return b
}

func writeInt64(b []byte, v int64) []byte {
	p := [maxVarintLen64]byte{}
	n := writeReverseVarint(p[:], v)
	off := maxVarintLen64 - n

	b = append(b, p[off:]...)
	b = append(b, byte(TypeInt64))
	return b
}

func writeUint8(b []byte, v uint8) []byte {
	b = append(b, v)
	b = append(b, byte(TypeUint8))
	return b
}

func writeUint16(b []byte, v uint16) []byte {
	p := [maxVarintLen16]byte{}
	n := writeReverseUvarint(p[:], uint64(v))
	off := maxVarintLen16 - n

	b = append(b, p[off:]...)
	b = append(b, byte(TypeUint16))
	return b
}

func writeUint32(b []byte, v uint32) []byte {
	p := [maxVarintLen32]byte{}
	n := writeReverseUvarint(p[:], uint64(v))
	off := maxVarintLen32 - n

	b = append(b, p[off:]...)
	b = append(b, byte(TypeUint32))
	return b
}

func writeUint64(b []byte, v uint64) []byte {
	p := [maxVarintLen64]byte{}
	n := writeReverseUvarint(p[:], v)
	off := maxVarintLen64 - n

	b = append(b, p[off:]...)
	b = append(b, byte(TypeUint64))
	return b
}

func writeFloat32(b []byte, v float32) []byte {
	p := [4]byte{}
	binary.BigEndian.PutUint32(p[:], math.Float32bits(v))

	b = append(b, p[:]...)
	b = append(b, byte(TypeFloat32))
	return b
}

func writeFloat64(b []byte, v float64) []byte {
	p := [8]byte{}
	binary.BigEndian.PutUint64(p[:], math.Float64bits(v))

	b = append(b, p[:]...)
	b = append(b, byte(TypeFloat64))
	return b
}

// bytes

func writeBytes(b []byte, v []byte) ([]byte, error) {
	size := len(v)
	if size > MaxSize {
		return nil, fmt.Errorf("write: bytes too large, max size=%d, actual size=%d", MaxSize, size)
	}

	b = append(b, v...)
	b = _writeBytesSize(b, uint32(size))
	b = append(b, byte(TypeBytes))
	return b, nil
}

func _writeBytesSize(b []byte, size uint32) []byte {
	p := [maxVarintLen32]byte{}
	n := writeReverseUvarint(p[:], uint64(size))
	off := maxVarintLen32 - n
	return append(b, p[off:]...)
}

// string

func writeString(b []byte, s string) ([]byte, error) {
	size := len(s)
	if size > MaxSize {
		return nil, fmt.Errorf("write: string too large, max size=%d, actual size=%d", MaxSize, size)
	}

	b = append(b, s...)
	b = append(b, 0) // zero byte
	b = _writeStringSize(b, uint32(size))
	b = append(b, byte(TypeString))
	return b, nil
}

func _writeStringSize(b []byte, size uint32) []byte {
	p := [maxVarintLen32]byte{}
	n := writeReverseUvarint(p[:], uint64(size))
	off := maxVarintLen32 - n
	return append(b, p[off:]...)
}

// list

func writeList(b []byte, bodySize int, table []listElement) ([]byte, error) {
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
	b, tsize, err = _writeListTable(b, table, big)
	if err != nil {
		return nil, err
	}

	// write sizes and type
	b = _writeListBodySize(b, bsize)
	b = _writeListTableSize(b, tsize)
	b = append(b, byte(type_))
	return b, nil
}

func _writeListTable(b []byte, table []listElement, big bool) ([]byte, uint32, error) {
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
	b, p := writeAlloc(b, size)
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

func _writeListTableSize(b []byte, size uint32) []byte {
	p := [maxVarintLen32]byte{}
	n := writeReverseUvarint(p[:], uint64(size))
	off := maxVarintLen32 - n
	return append(b, p[off:]...)
}

func _writeListBodySize(b []byte, size uint32) []byte {
	p := [maxVarintLen32]byte{}
	n := writeReverseUvarint(p[:], uint64(size))
	off := maxVarintLen32 - n
	return append(b, p[off:]...)
}

// message

func writeMessage(b []byte, bodySize int, table []messageField) ([]byte, error) {
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
	b, tsize, err = _writeMessageTable(b, table, big)
	if err != nil {
		return nil, err
	}

	// write sizes and type
	b = _writeMessageBodySize(b, bsize)
	b = _writeMessageTableSize(b, tsize)
	b = append(b, byte(type_))
	return b, nil
}

func _writeMessageTable(b []byte, table []messageField, big bool) ([]byte, uint32, error) {
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
	b, p := writeAlloc(b, size)
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

func _writeMessageTableSize(b []byte, size uint32) []byte {
	p := [maxVarintLen32]byte{}
	n := writeReverseUvarint(p[:], uint64(size))
	off := maxVarintLen32 - n
	return append(b, p[off:]...)
}

func _writeMessageBodySize(b []byte, size uint32) []byte {
	p := [maxVarintLen32]byte{}
	n := writeReverseUvarint(p[:], uint64(size))
	off := maxVarintLen32 - n
	return append(b, p[off:]...)
}

// private

// writeAlloc grows a buffer by n bytes and returns a new buffer and an allocated segment.
func writeAlloc(b []byte, n int) ([]byte, []byte) {
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
