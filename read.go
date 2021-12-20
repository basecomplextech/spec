package spec

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

func ReadType(b []byte) (Type, error) {
	t, n := _readType(b)
	if n < 0 {
		return 0, fmt.Errorf("read type: invalid data")
	}
	return t, nil
}

func ReadBool(b []byte) (bool, error) {
	t, n := _readType(b)
	if n < 0 {
		return false, fmt.Errorf("read bool: invalid data")
	}
	return t == TypeTrue, nil
}

func ReadByte(b []byte) (byte, error) {
	return ReadUInt8(b)
}

func ReadInt8(b []byte) (int8, error) {
	v, n := _readInt(b)
	switch {
	case n < 0:
		return 0, fmt.Errorf("read int8: invalid data")
	case v < math.MinInt8:
		return 0, fmt.Errorf("read int8: overflow, value too little")
	case v > math.MaxInt8:
		return 0, fmt.Errorf("read int8: overflow, value too large")
	}
	return int8(v), nil
}

func ReadInt16(b []byte) (int16, error) {
	v, n := _readInt(b)
	switch {
	case n < 0:
		return 0, fmt.Errorf("read int16: invalid data")
	case v < math.MinInt16:
		return 0, fmt.Errorf("read int16: overflow, value too little")
	case v > math.MaxInt16:
		return 0, fmt.Errorf("read int16: overflow, value too large")
	}
	return int16(v), nil
}

func ReadInt32(b []byte) (int32, error) {
	v, n := _readInt(b)
	switch {
	case n < 0:
		return 0, fmt.Errorf("read int32: invalid data")
	case v < math.MinInt32:
		return 0, fmt.Errorf("read int32: overflow, value too little")
	case v > math.MaxInt32:
		return 0, fmt.Errorf("read int32: overflow, value too large")
	}
	return int32(v), nil
}

func ReadInt64(b []byte) (int64, error) {
	v, n := _readInt(b)
	if n < 0 {
		return 0, fmt.Errorf("read int32: invalid data")
	}
	return v, nil
}

func ReadUInt8(b []byte) (uint8, error) {
	v, n := _readUint(b)
	switch {
	case n < 0:
		return 0, fmt.Errorf("read uint8: invalid data")
	case v > math.MaxUint8:
		return 0, fmt.Errorf("read uint8: overflow, value too large")
	}
	return uint8(v), nil
}

func ReadUInt16(b []byte) (uint16, error) {
	v, n := _readUint(b)
	switch {
	case n < 0:
		return 0, fmt.Errorf("read uint16: invalid data")
	case v > math.MaxUint16:
		return 0, fmt.Errorf("read uint16: overflow, value too large")
	}
	return uint16(v), nil
}

func ReadUInt32(b []byte) (uint32, error) {
	v, n := _readUint(b)
	switch {
	case n < 0:
		return 0, fmt.Errorf("read uint32: invalid data")
	case v > math.MaxUint32:
		return 0, fmt.Errorf("read uint32: overflow, value too large")
	}
	return uint32(v), nil
}

func ReadUInt64(b []byte) (uint64, error) {
	v, n := _readUint(b)
	if n < 0 {
		return 0, fmt.Errorf("read uint64: invalid data")
	}
	return v, nil
}

func ReadFloat32(b []byte) (float32, error) {
	v, n := _readFloat(b)
	switch {
	case n < 0:
		return 0, fmt.Errorf("read float32: invalid data")
	case v < math.SmallestNonzeroFloat32:
		return 0, fmt.Errorf("read float32: overflow, value too small")
	case v > math.MaxFloat64:
		return 0, fmt.Errorf("read float32: overflow, value too large")
	}
	return float32(v), nil
}

func ReadFloat64(b []byte) (float64, error) {
	v, n := _readFloat(b)
	if n < 0 {
		return 0, fmt.Errorf("read float64: invalid data")
	}
	return v, nil
}

// Bytes

func ReadBytes(b []byte) ([]byte, error) {
	t, n := _readType(b)
	switch {
	case n < 0:
		return nil, fmt.Errorf("read bytes: invalid data")
	case t == TypeNil:
		return nil, nil
	case t != TypeBytes:
		return nil, fmt.Errorf("read bytes: unexpected type, expected=%d, actual=%d", TypeBytes, t)
	}

	// bytes size
	off := len(b) - 1
	size, n1 := readReverseUvarint32(b[:off])
	if n1 < 0 {
		return nil, fmt.Errorf("read bytes: invalid size")
	}

	// bytes body
	off -= n1 - int(size)
	if off < 0 {
		return nil, fmt.Errorf("read bytes: invalid data")
	}

	v := b[off:]
	return v, nil
}

// String

func ReadString(b []byte) (string, error) {
	t, n := _readType(b)
	switch {
	case n < 0:
		return "", fmt.Errorf("read string: invalid data")
	case t == TypeNil:
		return "", nil
	case t != TypeString:
		return "", fmt.Errorf("read string: unexpected type, expected=%d, actual=%d", TypeString, t)
	}

	// string size
	off := len(b) - 1
	size, n1 := readReverseUvarint32(b[:off])
	if n1 < 0 {
		return "", fmt.Errorf("read string: invalid size")
	}

	// string body
	start := off - n1 - int(size) - 1 // zero byte
	if start < 0 {
		return "", fmt.Errorf("read string: invalid data")
	}
	end := start + int(size)

	p := b[start:end]
	s := *(*string)(unsafe.Pointer(&p))
	return s, nil
}

// List

func ReadList(b []byte) (List, error) {
	list := List{}

	t, n := _readType(b)
	switch {
	case n < 0:
		return list, fmt.Errorf("read list: invalid data")
	case t != TypeList:
		return list, fmt.Errorf("read list: unexpected type, expected=%d, actual=%d", TypeList, t)
	}
	off := len(b) - 1

	// table size
	tsize, tn := readReverseUvarint32(b[:off])
	if tn < 0 {
		return list, fmt.Errorf("read list: invalid table size")
	}
	off -= int(tn)

	// data size
	dsize, dn := readReverseUvarint32(b[:off])
	if dn < 0 {
		return list, fmt.Errorf("read list: invalid data size")
	}
	off -= int(dn)

	// element table
	table, err := _readListTable(b[:off], tsize)
	if err != nil {
		return list, err
	}
	off -= int(tsize)

	// element data
	data, err := _readListData(b[:off], dsize)
	if err != nil {
		return list, err
	}
	off -= int(dsize)

	// done
	list = List{
		buffer: b[off:],
		table:  table,
		data:   data,
	}
	return list, nil
}

func _readListTable(b []byte, size uint32) (listTable, error) {
	off := len(b) - int(size)
	switch {
	case off < 0:
		return nil, fmt.Errorf("read list: invalid table, array too small")
	case size%listElementSize != 0:
		return nil, fmt.Errorf("read list: invalid table, size not divisible by %d, size=%d",
			listElementSize, size)
	}

	p := b[off:]
	v := listTable(p)
	return v, nil
}

func _readListData(b []byte, size uint32) ([]byte, error) {
	off := len(b) - int(size)
	if off < 0 {
		return nil, fmt.Errorf("read list: invalid data")
	}
	return b[off:], nil
}

// Message

func ReadMessage(b []byte) (Message, error) {
	msg := Message{}

	t, n := _readType(b)
	switch {
	case n < 0:
		return msg, fmt.Errorf("read message: invalid data")
	case t != TypeMessage:
		return msg, fmt.Errorf("read message: unexpected type, expected=%d, actual=%d", TypeMessage, t)
	}
	off := len(b) - 1

	// table size
	tsize, tn := readReverseUvarint32(b[:off])
	if tn < 0 {
		return msg, fmt.Errorf("read message: invalid table size")
	}
	off -= int(tn)

	// data size
	dsize, dn := readReverseUvarint32(b[:off])
	if dn < 0 {
		return msg, fmt.Errorf("read message: invalid data size")
	}
	off -= int(dn)

	// field table
	table, err := _readMessageTable(b[:off], tsize)
	if err != nil {
		return msg, err
	}
	off -= int(tsize)

	// field data
	data, err := _readMessageData(b[:off], dsize)
	if err != nil {
		return msg, err
	}
	off -= int(dsize)

	// done
	msg = Message{
		buffer: b[off:],
		table:  table,
		data:   data,
	}
	return msg, nil
}

func _readMessageTable(b []byte, size uint32) (messageTable, error) {
	off := len(b) - int(size)
	switch {
	case off < 0:
		return nil, fmt.Errorf("read message: invalid table, array too small")
	case size%messageFieldSize != 0:
		return nil, fmt.Errorf("read message: invalid table, size not divisible by %d, size=%d",
			messageFieldSize, size)
	}

	p := b[off:]
	v := messageTable(p)
	return v, nil
}

func _readMessageData(b []byte, size uint32) ([]byte, error) {
	off := len(b) - int(size)
	if off < 0 {
		return nil, fmt.Errorf("read message: invalid data")
	}
	return b[off:], nil
}

// internal

func _readType(b []byte) (Type, int) {
	if len(b) == 0 {
		return TypeNil, 0
	}

	v := b[len(b)-1]
	return Type(v), 1
}

// _readInt reads and returns any int as int64 and the number of read bytes n, or -n on error.
func _readInt(b []byte) (int64, int) {
	t, n := _readType(b)
	switch {
	case n < 0:
		return 0, n
	case t == TypeNil:
		return 0, n
	}
	b = b[:len(b)-n]

	switch t {
	case TypeTrue:
		return 1, n
	case TypeFalse:
		return 0, n

	case TypeInt8, TypeUInt8:
		if len(b) < 1 {
			return 0, -1
		}

		v := b[len(b)-1]
		return int64(v), 2

	case TypeInt16, TypeInt32, TypeInt64:
		return readReverseVarint(b)

	case TypeUInt16, TypeUInt32, TypeUInt64:
		v, n := readReverseUvarint(b)
		if n < 0 {
			return 0, n
		}
		return int64(v), n
	}

	return 0, -1
}

// _readInt reads and returns any int as uint64 and the number of read bytes n, or -n on error.
func _readUint(b []byte) (uint64, int) {
	t, n := _readType(b)
	switch {
	case n < 0:
		return 0, n
	case t == TypeNil:
		return 0, n
	}
	b = b[:len(b)-n]

	switch t {
	case TypeTrue:
		return 1, 1
	case TypeFalse:
		return 0, 1

	case TypeInt8, TypeUInt8:
		if len(b) < 1 {
			return 0, -1
		}

		v := b[len(b)-1]
		return uint64(v), 2

	case TypeInt16, TypeInt32, TypeInt64:
		v, n := readReverseVarint(b)
		if n < 0 {
			return 0, n
		}
		return uint64(v), n

	case TypeUInt16, TypeUInt32, TypeUInt64:
		return readReverseUvarint(b)
	}

	return 0, -1
}

// _readInt reads and returns any float as float64 and the number of read bytes n, or -n on error.
func _readFloat(b []byte) (float64, int) {
	t, n := _readType(b)
	switch {
	case n < 0:
		return 0, n
	case t == TypeNil:
		return 0, n
	}

	switch t {
	case TypeFloat32:
		off := len(b) - 5
		if off < 0 {
			return 0, -1
		}

		v := binary.BigEndian.Uint32(b[off:])
		v1 := math.Float32frombits(v)
		return float64(v1), 5

	case TypeFloat64:
		off := len(b) - 9
		if off < 0 {
			return 0, -1
		}

		v := binary.BigEndian.Uint64(b[off:])
		v1 := math.Float64frombits(v)
		return v1, 9
	}

	return 0, -1
}
