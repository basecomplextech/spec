package spec

import (
	"encoding/binary"
	"math"
	"unsafe"
)

func ReadType(b []byte) (Type, int) {
	return readType(b)
}

func ReadBool(b []byte) (bool, int) {
	t, n := readType(b)
	if n < 0 {
		return false, n
	}
	return t == TypeTrue, 1
}

func ReadByte(b []byte) (byte, int) {
	return ReadUInt8(b)
}

func ReadInt8(b []byte) (int8, int) {
	if len(b) < 2 {
		return 0, -1
	}

	t, n := readType(b)
	switch {
	case n < 0:
		return 0, n
	case t != TypeInt8:
		return 0, -1
	}

	off := len(b) - 2
	v := b[off]
	return int8(v), 2
}

func ReadInt16(b []byte) (int16, int) {
	if len(b) == 0 {
		return 0, 0
	}

	t, n := readType(b)
	switch {
	case n < 0:
		return 0, n
	case t != TypeInt16:
		return 0, -1
	}

	off := len(b) - n
	v, n1 := readReverseVarint(b[:off])
	return int16(v), n + n1
}

func ReadInt32(b []byte) (int32, int) {
	if len(b) == 0 {
		return 0, 0
	}

	t, n := readType(b)
	switch {
	case n < 0:
		return 0, n
	case t != TypeInt32:
		return 0, -1
	}

	off := len(b) - n
	v, n1 := readReverseVarint(b[:off])
	return int32(v), n + n1
}

func ReadInt64(b []byte) (int64, int) {
	if len(b) == 0 {
		return 0, 0
	}

	t, n := readType(b)
	switch {
	case n < 0:
		return 0, n
	case t != TypeInt64:
		return 0, -1
	}

	off := len(b) - n
	v, n1 := readReverseVarint(b[:off])
	return v, n + n1
}

func ReadUInt8(b []byte) (uint8, int) {
	if len(b) < 2 {
		return 0, -1
	}

	t, n := readType(b)
	switch {
	case n < 0:
		return 0, n
	case t != TypeUInt8:
		return 0, -1
	}

	off := len(b) - 2
	return b[off], 2
}

func ReadUInt16(b []byte) (uint16, int) {
	if len(b) == 0 {
		return 0, 0
	}

	t, n := readType(b)
	switch {
	case n < 0:
		return 0, n
	case t != TypeUInt16:
		return 0, -1
	}

	off := len(b) - n
	v, n1 := readReverseUvarint(b[:off])
	return uint16(v), n + n1
}

func ReadUInt32(b []byte) (uint32, int) {
	if len(b) == 0 {
		return 0, 0
	}

	t, n := readType(b)
	switch {
	case n < 0:
		return 0, n
	case t != TypeUInt32:
		return 0, -1
	}

	off := len(b) - n
	v, n1 := readReverseUvarint(b[:off])
	return uint32(v), n + n1
}

func ReadUInt64(b []byte) (uint64, int) {
	if len(b) == 0 {
		return 0, 0
	}

	t, n := readType(b)
	switch {
	case n < 0:
		return 0, n
	case t != TypeUInt64:
		return 0, -1
	}

	off := len(b) - 1
	v, n1 := readReverseUvarint(b[:off])
	return v, n + n1
}

func ReadFloat32(b []byte) (float32, int) {
	if len(b) < 5 {
		return 0, -1
	}

	t, n := readType(b)
	switch {
	case n < 0:
		return 0, n
	case t != TypeFloat32:
		return 0, -1
	}

	off := len(b) - 5
	v := binary.BigEndian.Uint32(b[off:])
	v1 := math.Float32frombits(v)
	return v1, 5
}

func ReadFloat64(b []byte) (float64, int) {
	if len(b) < 9 {
		return 0, -1
	}

	t, n := readType(b)
	switch {
	case n < 0:
		return 0, n
	case t != TypeFloat64:
		return 0, -1
	}

	off := len(b) - 9
	v := binary.BigEndian.Uint64(b[off:])
	v1 := math.Float64frombits(v)
	return v1, 9
}

// Bytes

func ReadBytes(b []byte) ([]byte, int) {
	t, n := readType(b)
	switch {
	case n < 0:
		return nil, n
	case t != TypeBytes:
		return nil, -1
	}
	off := len(b) - 1

	// bytes size
	size, n1 := readReverseUvarint32(b[:off])
	off -= n1

	// bytes body
	v := readBytesBody(b[:off], size)
	return v, n + n1 + int(size)
}

func readBytesBody(b []byte, size uint32) []byte {
	if size == 0 {
		return nil
	}

	off := len(b) - int(size)
	if off < 0 {
		return nil
	}
	return b[off:]
}

// String

func ReadString(b []byte) (string, int) {
	t, n := readType(b)
	switch {
	case n < 0:
		return "", n
	case t != TypeString:
		return "", -1
	}
	off := len(b) - 1

	// string size
	size, n1 := readReverseUvarint32(b[:off])
	off -= n1

	// string body
	s := readStringBody(b[:off], size)
	return s, n + n1 + int(size)
}

func readStringBody(b []byte, size uint32) string {
	// mind last zero byte
	if size <= 1 {
		return ""
	}

	start := len(b) - int(size)
	if start < 0 {
		return ""
	}
	end := len(b) - 1 // last zero byte

	p := b[start:end]
	s := *(*string)(unsafe.Pointer(&p))
	return s
}

// List

func ReadList(b []byte) List {
	type_, _ := readType(b)
	if type_ != TypeList {
		return List{}
	}
	off := len(b) - 1

	// table size
	tsize, tn := readReverseUvarint32(b[:off])
	off -= int(tn)

	// data size
	dsize, dn := readReverseUvarint32(b[:off])
	off -= int(dn)

	// element table
	table := readListTable(b[:off], tsize)
	off -= int(tsize)

	// element data
	data := readListData(b[:off], dsize)
	off -= int(dsize)

	buffer := b[off:]
	return List{
		buffer: buffer,
		table:  table,
		data:   data,
	}
}

func readListTable(b []byte, size uint32) listTable {
	off := len(b) - int(size)
	if off < 0 {
		return nil
	}

	p := b[off:]
	v := listTable(p)
	return v
}

func readListData(b []byte, size uint32) []byte {
	off := len(b) - int(size)
	if off < 0 {
		return nil
	}
	return b[off:]
}

// Message

func ReadMessage(b []byte) Message {
	type_, _ := readType(b)
	if type_ != TypeMessage {
		return Message{}
	}
	off := len(b) - 1

	// table size
	tsize, tn := readReverseUvarint32(b[:off])
	off -= int(tn)

	// data size
	dsize, dn := readReverseUvarint32(b[:off])
	off -= int(dn)

	// field table
	table := readMessageTable(b[:off], tsize)
	off -= int(tsize)

	// field data
	data := readMessageData(b[:off], dsize)
	off -= int(dsize)

	buffer := b[off:]
	return Message{
		buffer: buffer,
		table:  table,
		data:   data,
	}
}

func readMessageTable(b []byte, size uint32) messageTable {
	off := len(b) - int(size)
	if off < 0 {
		return nil
	}

	p := b[off:]
	return messageTable(p)
}

func readMessageData(b []byte, size uint32) []byte {
	off := len(b) - int(size)
	if off < 0 {
		return nil
	}

	return b[off:]
}

// internal

func readType(b []byte) (Type, int) {
	if len(b) == 0 {
		return TypeNil, 1
	}

	off := len(b) - 1
	v := b[off]
	return Type(v), 1
}
