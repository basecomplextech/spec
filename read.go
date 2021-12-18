package spec

import (
	"encoding/binary"
	"math"
	"unsafe"
)

func ReadType(b []byte) Type {
	t, _ := readType(b)
	return t
}

func ReadBool(b []byte) bool {
	t, _ := readType(b)
	if t != TypeTrue {
		return false
	}
	return true
}

func ReadByte(b []byte) byte {
	return ReadUInt8(b)
}

func ReadInt8(b []byte) int8 {
	if len(b) < 2 {
		return 0
	}

	off := len(b) - 1
	t := Type(b[off])
	if t != TypeInt8 {
		return 0
	}

	v := b[off-1]
	return int8(v)
}

func ReadInt16(b []byte) (int16, int) {
	if len(b) == 0 {
		return 0, 0
	}

	off := len(b) - 1
	t := Type(b[off])
	if t != TypeInt16 {
		return 0, -1
	}

	v, rem := reverseVarint(b[:off])
	return int16(v), rem
}

func ReadInt32(b []byte) (int32, int) {
	if len(b) == 0 {
		return 0, 0
	}

	off := len(b) - 1
	t := Type(b[off])
	if t != TypeInt32 {
		return 0, -1
	}

	v, rem := reverseVarint(b[:off])
	return int32(v), rem
}

func ReadInt64(b []byte) (int64, int) {
	if len(b) == 0 {
		return 0, 0
	}

	off := len(b) - 1
	t := Type(b[off])
	if t != TypeInt64 {
		return 0, -1
	}

	v, rem := reverseVarint(b[:off])
	return v, rem
}

func ReadUInt8(b []byte) uint8 {
	if len(b) < 2 {
		return 0
	}

	off := len(b) - 1
	t := Type(b[off])
	if t != TypeUInt8 {
		return 0
	}

	return b[off-1]
}

func ReadUInt16(b []byte) (uint16, int) {
	if len(b) == 0 {
		return 0, 0
	}

	off := len(b) - 1
	t := Type(b[off])
	if t != TypeUInt16 {
		return 0, -1
	}

	v, rem := reverseUvarint(b[:off])
	return uint16(v), rem
}

func ReadUInt32(b []byte) (uint32, int) {
	if len(b) == 0 {
		return 0, 0
	}

	off := len(b) - 1
	t := Type(b[off])
	if t != TypeUInt32 {
		return 0, -1
	}

	v, rem := reverseUvarint(b[:off])
	return uint32(v), rem
}

func ReadUInt64(b []byte) (uint64, int) {
	if len(b) == 0 {
		return 0, 0
	}

	off := len(b) - 1
	t := Type(b[off])
	if t != TypeUInt64 {
		return 0, -1
	}

	v, rem := reverseUvarint(b[:off])
	return v, rem
}

func ReadFloat32(b []byte) float32 {
	if len(b) < 5 {
		return 0
	}

	off := len(b) - 1
	t := Type(b[off])
	if t != TypeFloat32 {
		return 0
	}

	off -= 4
	v := binary.BigEndian.Uint32(b[off:])
	return math.Float32frombits(v)
}

func ReadFloat64(b []byte) float64 {
	if len(b) < 9 {
		return 0
	}

	off := len(b) - 1
	t := Type(b[off])
	if t != TypeFloat64 {
		return 0
	}

	off -= 8
	v := binary.BigEndian.Uint64(b[off:])
	return math.Float64frombits(v)
}

func ReadBytes(b []byte) []byte {
	t, b := readType(b)
	if t != TypeBytes {
		return nil
	}

	size, b := readBytesSize(b)
	return readBytesBody(b, size)
}

func ReadString(b []byte) string {
	t, b := readType(b)
	if t != TypeString {
		return ""
	}

	size, b := readStringSize(b)
	return readStringBody(b, size)
}

// internal

func readType(b []byte) (Type, []byte) {
	if len(b) == 0 {
		return TypeNil, nil
	}

	off := len(b) - 1
	v := b[off]
	return Type(v), b[:off]
}

// bytes

func readBytesSize(b []byte) (uint32, []byte) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
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

// string

func readStringSize(b []byte) (uint32, []byte) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func readStringBody(b []byte, size uint32) string {
	// mind last zero byte
	if size <= 1 {
		return ""
	}

	off := len(b) - int(size)
	if off < 0 {
		return ""
	}

	end := off + int(size) - 1
	p := b[off:end]
	s := *(*string)(unsafe.Pointer(&p))
	return s
}

// list

func readListBuffer(b []byte, tableSize uint32, dataSize uint32) []byte {
	size := int(1 + 4 + 4 + tableSize + dataSize) // type(1) + table size (4) + data size (4)
	off := len(b) - size
	if off < 0 {
		return nil
	}

	return b[off:]
}

func readListTable(b []byte, size uint32) (listTable, []byte) {
	off := len(b) - int(size)
	if off < 0 {
		return nil, nil
	}

	p := b[off:]
	v := listTable(p)
	return v, b[:off]
}

func readListTableSize(b []byte) (uint32, []byte) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func readListDataSize(b []byte) (uint32, []byte) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

// message

func readMessageBuffer(b []byte, tableSize uint32, dataSize uint32) []byte {
	size := 1 + 4 + 4 + int(tableSize) + int(dataSize) // type(1) + table size(4) + data size(4)
	off := len(b) - size
	if off < 0 {
		return nil
	}

	return b[off:]
}

func readMessageTable(b []byte, size uint32) (messageTable, []byte) {
	off := len(b) - int(size)
	if off < 0 {
		return nil, nil
	}

	p := b[off:]
	v := messageTable(p)
	return v, b[:off]
}

func readMessageTableSize(b []byte) (uint32, []byte) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}

func readMessageDataSize(b []byte) (uint32, []byte) {
	off := len(b) - 4
	if off < 0 {
		return 0, nil
	}

	p := b[off:]
	v := binary.BigEndian.Uint32(p)
	return v, b[:off]
}
