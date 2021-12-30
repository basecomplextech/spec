package spec

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

func readType(b []byte) (Type, error) {
	t, n := _readType(b)
	if n < 0 {
		return 0, fmt.Errorf("read type: invalid data")
	}
	return t, nil
}

func _readType(b []byte) (Type, int) {
	if len(b) == 0 {
		return TypeNil, 0
	}

	v := b[len(b)-1]
	return Type(v), 1
}

// bool

func readBool(b []byte) (bool, error) {
	t, n := _readType(b)
	if n < 0 {
		return false, fmt.Errorf("read bool: invalid data")
	}
	return t == TypeTrue, nil
}

// int

func readInt8(b []byte) (int8, error) {
	v, n := _readInt(b)
	switch {
	case n < 0:
		return 0, fmt.Errorf("read int8: invalid data")
	case v < math.MinInt8:
		return 0, fmt.Errorf("read int8: overflow, value too small")
	case v > math.MaxInt8:
		return 0, fmt.Errorf("read int8: overflow, value too large")
	}
	return int8(v), nil
}

func readInt16(b []byte) (int16, error) {
	v, n := _readInt(b)
	switch {
	case n < 0:
		return 0, fmt.Errorf("read int16: invalid data")
	case v < math.MinInt16:
		return 0, fmt.Errorf("read int16: overflow, value too small")
	case v > math.MaxInt16:
		return 0, fmt.Errorf("read int16: overflow, value too large")
	}
	return int16(v), nil
}

func readInt32(b []byte) (int32, error) {
	v, n := _readInt(b)
	switch {
	case n < 0:
		return 0, fmt.Errorf("read int32: invalid data")
	case v < math.MinInt32:
		return 0, fmt.Errorf("read int32: overflow, value too small")
	case v > math.MaxInt32:
		return 0, fmt.Errorf("read int32: overflow, value too large")
	}
	return int32(v), nil
}

func readInt64(b []byte) (int64, error) {
	v, n := _readInt(b)
	if n < 0 {
		return 0, fmt.Errorf("read int32: invalid data")
	}
	return v, nil
}

// _readInt reads and returns any int as int64 and the number of read bytes n, or -n on error.
func _readInt(b []byte) (int64, int) {
	// type
	t, n := _readType(b)
	switch {
	case n < 0:
		return 0, n
	case t == TypeNil:
		return 0, n
	}
	off := len(b) - n
	b = b[:off]

	// read, cast int
	switch t {
	case TypeTrue:
		return 1, n
	case TypeFalse:
		return 0, n

	case TypeInt8,
		TypeUint8:
		if len(b) < 1 {
			return 0, -1
		}
		v := b[len(b)-1]
		return int64(v), n + 1

	case TypeInt16,
		TypeInt32,
		TypeInt64:
		v, vn := readReverseVarint(b)
		if vn < 0 {
			return 0, vn
		}
		return v, n + vn

	case TypeUint16,
		TypeUint32,
		TypeUint64:
		v, vn := readReverseUvarint(b)
		if vn < 0 {
			return 0, vn
		}
		return int64(v), n + vn
	}

	return 0, -1
}

// uint

func readByte(b []byte) (byte, error) {
	return readUint8(b)
}

func readUint8(b []byte) (uint8, error) {
	v, n := _readUint(b)
	switch {
	case n < 0:
		return 0, fmt.Errorf("read uint8: invalid data")
	case v > math.MaxUint8:
		return 0, fmt.Errorf("read uint8: overflow, value too large")
	}
	return uint8(v), nil
}

func readUint16(b []byte) (uint16, error) {
	v, n := _readUint(b)
	switch {
	case n < 0:
		return 0, fmt.Errorf("read uint16: invalid data")
	case v > math.MaxUint16:
		return 0, fmt.Errorf("read uint16: overflow, value too large")
	}
	return uint16(v), nil
}

func readUint32(b []byte) (uint32, error) {
	v, n := _readUint(b)
	switch {
	case n < 0:
		return 0, fmt.Errorf("read uint32: invalid data")
	case v > math.MaxUint32:
		return 0, fmt.Errorf("read uint32: overflow, value too large")
	}
	return uint32(v), nil
}

func readUint64(b []byte) (uint64, error) {
	v, n := _readUint(b)
	if n < 0 {
		return 0, fmt.Errorf("read uint64: invalid data")
	}
	return v, nil
}

// _readUint reads and returns any int as uint64 and the number of read bytes n, or -n on error.
func _readUint(b []byte) (uint64, int) {
	// type
	t, n := _readType(b)
	switch {
	case n < 0:
		return 0, n
	case t == TypeNil:
		return 0, n
	}
	off := len(b) - n
	b = b[:off]

	// read, cast uint
	switch t {
	case TypeTrue:
		return 1, n
	case TypeFalse:
		return 0, n

	case TypeInt8,
		TypeUint8:
		if len(b) < 1 {
			return 0, -1
		}
		v := b[len(b)-1]
		return uint64(v), n + 1

	case TypeInt16,
		TypeInt32,
		TypeInt64:
		v, vn := readReverseVarint(b)
		if vn < 0 {
			return 0, vn
		}
		return uint64(v), n + vn

	case TypeUint16,
		TypeUint32,
		TypeUint64:
		v, vn := readReverseUvarint(b)
		if vn < 0 {
			return 0, vn
		}
		return v, n + vn
	}

	return 0, -1
}

// float

func readFloat32(b []byte) (float32, error) {
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

func readFloat64(b []byte) (float64, error) {
	v, n := _readFloat(b)
	if n < 0 {
		return 0, fmt.Errorf("read float64: invalid data")
	}
	return v, nil
}

// _readFloat reads and returns any float as float64 and the number of read bytes n, or -n on error.
func _readFloat(b []byte) (float64, int) {
	// type
	t, n := _readType(b)
	switch {
	case n < 0:
		return 0, n
	case t == TypeNil:
		return 0, n
	}

	// read, cast float
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

// bytes

func readBytes(b []byte) ([]byte, error) {
	// type
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
	off := len(b) - n
	size, sn := _readBytesSize(b[:off])
	if sn < 0 {
		return nil, fmt.Errorf("read bytes: invalid size")
	}

	// bytes body
	off -= sn
	return _readBytesBody(b[:off], size)
}

func _readBytesSize(b []byte) (uint32, int) {
	return readReverseUvarint32(b)
}

func _readBytesBody(b []byte, size uint32) ([]byte, error) {
	end := len(b)
	start := end - int(size)
	if start < 0 {
		return nil, fmt.Errorf("read bytes: invalid body, expected size=%d, actual size=%d", size, len(b))
	}

	v := b[start:end]
	return v, nil
}

// string

func readString(b []byte) (string, error) {
	// type
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
	off := len(b) - n
	size, sn := _readStringSize(b[:off])
	if sn < 0 {
		return "", fmt.Errorf("read string: invalid size")
	}

	// string body
	off -= (sn + 1) // zero byte
	return _readStringBody(b[:off], size)
}

func _readStringSize(b []byte) (uint32, int) {
	return readReverseUvarint32(b)
}

func _readStringBody(b []byte, size uint32) (string, error) {
	end := len(b)
	start := end - int(size)
	if start < 0 {
		return "", fmt.Errorf("read string: invalid body, expected size=%d, actual size=%d", size, len(b))
	}

	p := b[start:end]
	s := *(*string)(unsafe.Pointer(&p))
	return s, nil
}

// list

func readList(b []byte) (list, error) {
	l := list{}
	if len(b) == 0 {
		return l, nil
	}

	// read type
	t, n := _readType(b)
	if n < 0 {
		return l, fmt.Errorf("read list: invalid data")
	}

	// check type
	switch t {
	default:
		return l, fmt.Errorf("read list: unexpected type, expected=%d, actual=%d", TypeList, t)
	case TypeNil:
		return l, nil
	case TypeList, TypeListBig:
	}
	big := t == TypeListBig

	// table size
	off := len(b) - 1
	tsize, tn := _readListTableSize(b[:off])
	if tn < 0 {
		return l, fmt.Errorf("read list: invalid table size")
	}

	// body size
	off -= int(tn)
	bsize, dn := _readListBodySize(b[:off])
	if dn < 0 {
		return l, fmt.Errorf("read list: invalid body size")
	}

	// table
	off -= int(dn)
	table, err := _readListTable(b[:off], tsize, big)
	if err != nil {
		return l, err
	}

	// body
	off -= int(tsize)
	off -= int(bsize)
	if off < 0 {
		return l, fmt.Errorf("read list: invalid body")
	}

	// done
	l = list{
		data:  b[off:],
		table: table,
		body:  bsize,
		big:   big,
	}
	return l, nil
}

func _readListTableSize(b []byte) (uint32, int) {
	return readReverseUvarint32(b)
}

func _readListBodySize(b []byte) (uint32, int) {
	return readReverseUvarint32(b)
}

func _readListTable(b []byte, size uint32, big bool) (listTable, error) {
	// element size
	var elemSize uint32
	if big {
		elemSize = listElementBigSize
	} else {
		elemSize = listElementSmallSize
	}

	// check offset
	off := len(b) - int(size)
	if off < 0 {
		return nil, fmt.Errorf("read list: invalid table, array too small")
	}

	// check divisible
	if size%elemSize != 0 {
		return nil, fmt.Errorf("read list: invalid table, size not divisible by %d, size=%d",
			elemSize, size)
	}

	p := b[off:]
	v := listTable(p)
	return v, nil
}

// message

func readMessage(b []byte) (message, error) {
	msg := message{}
	if len(b) == 0 {
		return msg, nil
	}

	// read type
	t, n := _readType(b)
	if n < 0 {
		return msg, fmt.Errorf("read message: invalid data")
	}

	// check type
	switch t {
	default:
		return msg, fmt.Errorf("read message: unexpected type, expected=%d, actual=%d", TypeMessage, t)
	case TypeNil:
		return msg, nil
	case TypeMessage, TypeMessageBig:
	}
	big := t == TypeMessageBig

	// table size
	off := len(b) - 1
	tsize, tn := _readMessageTableSize(b[:off])
	if tn < 0 {
		return msg, fmt.Errorf("read message: invalid table size")
	}

	// body size
	off -= int(tn)
	bsize, dn := _readMessageBodySize(b[:off])
	if dn < 0 {
		return msg, fmt.Errorf("read message: invalid body size")
	}

	// table
	off -= int(dn)
	table, err := _readMessageTable(b[:off], tsize, big)
	if err != nil {
		return msg, err
	}

	// body
	off -= int(tsize)
	off -= int(bsize)
	if off < 0 {
		return msg, fmt.Errorf("read message: invalid body")
	}

	// done
	msg = message{
		data:  b[off:],
		table: table,
		body:  bsize,
		big:   big,
	}
	return msg, nil
}

func _readMessageTableSize(b []byte) (uint32, int) {
	return readReverseUvarint32(b)
}

func _readMessageBodySize(b []byte) (uint32, int) {
	return readReverseUvarint32(b)
}

func _readMessageTable(b []byte, size uint32, big bool) (messageTable, error) {
	// field size
	var fieldSize uint32
	if big {
		fieldSize = messageFieldBigSize
	} else {
		fieldSize = messageFieldSmallSize
	}

	// check offset
	off := len(b) - int(size)
	if off < 0 {
		return nil, fmt.Errorf("read message: invalid table, array too small")
	}

	// check divisible
	if size%fieldSize != 0 {
		return nil, fmt.Errorf("read message: invalid table, size not divisible by %d, size=%d",
			fieldSize, size)
	}

	p := b[off:]
	v := messageTable(p)
	return v, nil
}
