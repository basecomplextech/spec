package spec

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"

	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
)

func readType(b []byte) (Type, int, error) {
	t, n := _readType(b)
	if n < 0 {
		return 0, -1, fmt.Errorf("read type: invalid data")
	}
	return t, n, nil
}

func _readType(b []byte) (Type, int) {
	if len(b) == 0 {
		return TypeNil, 0
	}

	v := b[len(b)-1]
	return Type(v), 1
}

// bool

func readBool(b []byte) (bool, int, error) {
	t, n := _readType(b)
	if n < 0 {
		return false, -1, fmt.Errorf("read bool: invalid data")
	}
	v := t == TypeTrue
	return v, n, nil
}

// int

func readInt8(b []byte) (int8, int, error) {
	v, n := _readInt(b)
	switch {
	case n < 0:
		return 0, -1, fmt.Errorf("read int8: invalid data")
	case v < math.MinInt8:
		return 0, -1, fmt.Errorf("read int8: overflow, value too small")
	case v > math.MaxInt8:
		return 0, -1, fmt.Errorf("read int8: overflow, value too large")
	}
	return int8(v), n, nil
}

func readInt16(b []byte) (int16, int, error) {
	v, n := _readInt(b)
	switch {
	case n < 0:
		return 0, -1, fmt.Errorf("read int16: invalid data")
	case v < math.MinInt16:
		return 0, -1, fmt.Errorf("read int16: overflow, value too small")
	case v > math.MaxInt16:
		return 0, -1, fmt.Errorf("read int16: overflow, value too large")
	}
	return int16(v), n, nil
}

func readInt32(b []byte) (int32, int, error) {
	v, n := _readInt(b)
	switch {
	case n < 0:
		return 0, -1, fmt.Errorf("read int32: invalid data")
	case v < math.MinInt32:
		return 0, -1, fmt.Errorf("read int32: overflow, value too small")
	case v > math.MaxInt32:
		return 0, -1, fmt.Errorf("read int32: overflow, value too large")
	}
	return int32(v), n, nil
}

func readInt64(b []byte) (int64, int, error) {
	v, n := _readInt(b)
	if n < 0 {
		return 0, n, fmt.Errorf("read int32: invalid data")
	}
	return v, n, nil
}

// _readInt reads and returns any int as int64 and the number of read bytes n, or -1 on error.
func _readInt(b []byte) (int64, int) {
	// type
	t, n := _readType(b)
	switch {
	case n < 0:
		return 0, -1
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
			return 0, -1
		}

		total := n + vn
		return v, total

	case TypeUint16,
		TypeUint32,
		TypeUint64:
		v, vn := readReverseUvarint(b)
		if vn < 0 {
			return 0, -1
		}

		total := n + vn
		return int64(v), total
	}

	return 0, -1
}

// uint

func readByte(b []byte) (byte, int, error) {
	return readUint8(b)
}

func readUint8(b []byte) (uint8, int, error) {
	v, n := _readUint(b)
	switch {
	case n < 0:
		return 0, -1, fmt.Errorf("read uint8: invalid data")
	case v > math.MaxUint8:
		return 0, -1, fmt.Errorf("read uint8: overflow, value too large")
	}
	return uint8(v), n, nil
}

func readUint16(b []byte) (uint16, int, error) {
	v, n := _readUint(b)
	switch {
	case n < 0:
		return 0, -1, fmt.Errorf("read uint16: invalid data")
	case v > math.MaxUint16:
		return 0, -1, fmt.Errorf("read uint16: overflow, value too large")
	}
	return uint16(v), n, nil
}

func readUint32(b []byte) (uint32, int, error) {
	v, n := _readUint(b)
	switch {
	case n < 0:
		return 0, -1, fmt.Errorf("read uint32: invalid data")
	case v > math.MaxUint32:
		return 0, -1, fmt.Errorf("read uint32: overflow, value too large")
	}
	return uint32(v), n, nil
}

func readUint64(b []byte) (uint64, int, error) {
	v, n := _readUint(b)
	if n < 0 {
		return 0, n, fmt.Errorf("read uint64: invalid data")
	}
	return v, n, nil
}

// _readUint reads and returns any int as uint64 and the number of read bytes n, or -n on error.
func _readUint(b []byte) (uint64, int) {
	// type
	t, n := _readType(b)
	switch {
	case n < 0:
		return 0, -1
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
			return 0, -1
		}

		total := n + vn
		return uint64(v), total

	case TypeUint16,
		TypeUint32,
		TypeUint64:
		v, vn := readReverseUvarint(b)
		if vn < 0 {
			return 0, -1
		}

		total := n + vn
		return v, total
	}

	return 0, -1
}

// u128/u256

func readU128(b []byte) (u128.U128, int, error) {
	// type
	t, n := _readType(b)
	switch {
	case n < 0:
		return u128.U128{}, -1, fmt.Errorf("read u128: invalid data")
	case t == TypeNil:
		return u128.U128{}, n, nil
	case t != TypeU128:
		return u128.U128{}, -1, fmt.Errorf("read u128: unexpected type, expected=%d, actual=%d", TypeU128, t)
	}

	end := len(b) - n
	start := end - 16
	if start < 0 {
		return u128.U128{}, -1, fmt.Errorf("read u128: invalid data")
	}

	p := b[start:end]
	v, err := u128.Parse(p)
	if err != nil {
		return u128.U128{}, -1, err
	}

	total := n + 16
	return v, total, nil
}

func readU256(b []byte) (u256.U256, int, error) {
	// type
	t, n := _readType(b)
	switch {
	case n < 0:
		return u256.U256{}, -1, fmt.Errorf("read u256: invalid data")
	case t == TypeNil:
		return u256.U256{}, -1, nil
	case t != TypeU256:
		return u256.U256{}, -1, fmt.Errorf("read u256: unexpected type, expected=%d, actual=%d", TypeU256, t)
	}

	end := len(b) - n
	start := end - 32
	if start < 0 {
		return u256.U256{}, -1, fmt.Errorf("read u256: invalid data")
	}

	p := b[start:end]
	v, err := u256.Parse(p)
	if err != nil {
		return u256.U256{}, -1, err
	}

	total := n + 32
	return v, total, err
}

// float

func readFloat32(b []byte) (float32, int, error) {
	v, n := _readFloat(b)
	switch {
	case n < 0:
		return 0, -1, fmt.Errorf("read float32: invalid data")
	case v < math.SmallestNonzeroFloat32:
		return 0, -1, fmt.Errorf("read float32: overflow, value too small")
	case v > math.MaxFloat64:
		return 0, -1, fmt.Errorf("read float32: overflow, value too large")
	}
	return float32(v), n, nil
}

func readFloat64(b []byte) (float64, int, error) {
	v, n := _readFloat(b)
	if n < 0 {
		return 0, n, fmt.Errorf("read float64: invalid data")
	}
	return v, n, nil
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

func readBytes(b []byte) ([]byte, int, error) {
	// type
	t, n := _readType(b)
	switch {
	case n < 0:
		return nil, -1, fmt.Errorf("read bytes: invalid data")
	case t == TypeNil:
		return nil, n, nil
	case t != TypeBytes:
		return nil, -1, fmt.Errorf("read bytes: unexpected type, expected=%d, actual=%d", TypeBytes, t)
	}

	// bytes size
	off := len(b) - n
	size, sn := _readBytesSize(b[:off])
	if sn < 0 {
		return nil, -1, fmt.Errorf("read bytes: invalid size")
	}

	// bytes body
	off -= sn
	body, err := _readBytesBody(b[:off], size)
	if err != nil {
		return nil, -1, err
	}

	total := n + sn + int(size)
	return body, total, nil
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

func readString(b []byte) (string, int, error) {
	// type
	t, n := _readType(b)
	switch {
	case n < 0:
		return "", -1, fmt.Errorf("read string: invalid data")
	case t == TypeNil:
		return "", n, nil
	case t != TypeString:
		return "", -1, fmt.Errorf("read string: unexpected type, expected=%d, actual=%d", TypeString, t)
	}

	// string size
	off := len(b) - n
	size, sn := _readStringSize(b[:off])
	if sn < 0 {
		return "", -1, fmt.Errorf("read string: invalid size")
	}

	// string body
	off -= (sn + 1) // zero byte
	body, err := _readStringBody(b[:off], size)
	if err != nil {
		return "", -1, err
	}

	total := n + sn + 1 + int(size)
	return body, total, err
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

func readList(b []byte) (List, int, error) {
	l := List{}
	if len(b) == 0 {
		return l, 0, nil
	}

	// read type
	t, n := _readType(b)
	if n < 0 {
		return l, -1, fmt.Errorf("read list: invalid data")
	}

	// check type
	switch t {
	default:
		return l, -1, fmt.Errorf("read list: unexpected type, expected=%d, actual=%d", TypeList, t)
	case TypeNil:
		return l, n, nil
	case TypeList, TypeListBig:
	}
	big := t == TypeListBig

	// table size
	off := len(b) - 1
	tsize, tn := _readListTableSize(b[:off])
	if tn < 0 {
		return l, -1, fmt.Errorf("read list: invalid table size")
	}

	// body size
	off -= int(tn)
	bsize, dn := _readListBodySize(b[:off])
	if dn < 0 {
		return l, -1, fmt.Errorf("read list: invalid body size")
	}

	// table
	off -= int(dn)
	table, err := _readListTable(b[:off], tsize, big)
	if err != nil {
		return l, -1, err
	}

	// body
	off -= int(tsize)
	off -= int(bsize)
	if off < 0 {
		return l, -1, fmt.Errorf("read list: invalid body")
	}

	// done
	l = List{
		data:  b[off:],
		table: table,
		body:  bsize,
		big:   big,
	}

	total := n + tn + dn + int(tsize) + int(bsize)
	return l, total, nil
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

func readMessage(b []byte) (Message, int, error) {
	msg := Message{}
	if len(b) == 0 {
		return msg, 0, nil
	}

	// read type
	t, n := _readType(b)
	if n < 0 {
		return msg, -1, fmt.Errorf("read message: invalid type")
	}

	// check type
	switch t {
	default:
		return msg, -1, fmt.Errorf("read message: unexpected type, expected=%d, actual=%d", TypeMessage, t)
	case TypeNil:
		return msg, 0, nil
	case TypeMessage, TypeMessageBig:
	}
	big := t == TypeMessageBig

	// table size
	off := len(b) - n
	tsize, tn := _readMessageTableSize(b[:off])
	if tn < 0 {
		return msg, -1, fmt.Errorf("read message: invalid table size")
	}

	// body size
	off -= int(tn)
	bsize, dn := _readMessageBodySize(b[:off])
	if dn < 0 {
		return msg, -1, fmt.Errorf("read message: invalid body size")
	}

	// table
	off -= int(dn)
	table, err := _readMessageTable(b[:off], tsize, big)
	if err != nil {
		return msg, -1, err
	}

	// body
	off -= int(tsize)
	off -= int(bsize)
	if off < 0 {
		return msg, -1, fmt.Errorf("read message: invalid body")
	}

	// done
	msg = Message{
		data:  b[off:],
		table: table,
		body:  bsize,
		big:   big,
	}

	total := n + tn + dn + int(tsize) + int(bsize)
	return msg, total, nil
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

// struct

func readStruct(b []byte) (bodySize int, n int, err error) {
	if len(b) == 0 {
		return 0, 0, nil
	}

	// read type
	t, n := _readType(b)
	if n < 0 {
		return 0, -1, fmt.Errorf("read struct: invalid type")
	}

	// check type
	switch t {
	default:
		return 0, -1, fmt.Errorf("read struct: unexpected type, expected=%d, actual=%d", TypeStruct, t)
	case TypeNil:
		return 0, 0, nil
	case TypeStruct:
	}

	// body size
	off := len(b) - n
	bsize, bn := _readStructBodySize(b[:off])
	if bn < 0 {
		return 0, -1, fmt.Errorf("read struct: invalid body size")
	}

	// done
	total := n + bn + int(bsize)
	return int(bsize), total, nil
}

func _readStructBodySize(b []byte) (uint32, int) {
	return readReverseUvarint32(b)
}
