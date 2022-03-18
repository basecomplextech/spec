package spec

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"unsafe"

	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
	"github.com/complexl/spec/rvarint"
)

func DecodeType(b []byte) (Type, int, error) {
	t, n := decodeType(b)
	if n < 0 {
		return 0, -1, fmt.Errorf("decode type: invalid data")
	}
	return t, n, nil
}

func decodeType(b []byte) (Type, int) {
	if len(b) == 0 {
		return TypeNil, 0
	}

	v := b[len(b)-1]
	return Type(v), 1
}

// Byte

func DecodeByte(b []byte) (byte, int, error) {
	v, n := decodeInt64(b)
	switch {
	case n < 0:
		return 0, -1, errors.New("decode byte: invalid data")
	case v < math.MinInt8:
		return 0, -1, errors.New("decode byte: overflow, value too small")
	case v > math.MaxInt8:
		return 0, -1, errors.New("decode byte: overflow, value too large")
	}
	return byte(v), n, nil
}

// Bool

func DecodeBool(b []byte) (bool, int, error) {
	t, n := decodeType(b)
	if n < 0 {
		return false, -1, errors.New("decode bool: invalid data")
	}
	v := t == TypeTrue
	return v, n, nil
}

// Int

func DecodeInt32(b []byte) (int32, int, error) {
	v, n := decodeInt64(b)
	switch {
	case n < 0:
		return 0, -1, errors.New("decode int32: invalid data")
	case v < math.MinInt32:
		return 0, -1, errors.New("decode int32: overflow, value too small")
	case v > math.MaxInt32:
		return 0, -1, errors.New("decode int32: overflow, value too large")
	}
	return int32(v), n, nil
}

func DecodeInt64(b []byte) (int64, int, error) {
	v, n := decodeInt64(b)
	if n < 0 {
		return 0, n, errors.New("decode int32: invalid data")
	}
	return v, n, nil
}

// decodeInt64 reads and returns any int as int64 and the number of decode bytes n, or -1 on error.
func decodeInt64(b []byte) (int64, int) {
	// type
	t, n := decodeType(b)
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

	case TypeByte:
		if len(b) < 1 {
			return 0, -1
		}
		v := b[len(b)-1]
		return int64(v), n + 1

	case TypeInt32,
		TypeInt64:
		v, vn := rvarint.Int64(b)
		if vn < 0 {
			return 0, -1
		}

		total := n + vn
		return v, total

	case TypeUint32,
		TypeUint64:
		v, vn := rvarint.Uint64(b)
		if vn < 0 {
			return 0, -1
		}

		total := n + vn
		return int64(v), total
	}

	return 0, -1
}

// Uint

func DecodeUint32(b []byte) (uint32, int, error) {
	v, n := decodeUint64(b)
	switch {
	case n < 0:
		return 0, -1, errors.New("decode uint32: invalid data")
	case v > math.MaxUint32:
		return 0, -1, errors.New("decode uint32: overflow, value too large")
	}
	return uint32(v), n, nil
}

func DecodeUint64(b []byte) (uint64, int, error) {
	v, n := decodeUint64(b)
	if n < 0 {
		return 0, n, fmt.Errorf("decode uint64: invalid data")
	}
	return v, n, nil
}

// decodeUint64 reads and returns any int as uint64 and the number of decode bytes n, or -n on error.
func decodeUint64(b []byte) (uint64, int) {
	// type
	t, n := decodeType(b)
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

	case TypeByte:
		if len(b) < 1 {
			return 0, -1
		}
		v := b[len(b)-1]
		return uint64(v), n + 1

	case TypeInt32,
		TypeInt64:
		v, vn := rvarint.Int64(b)
		if vn < 0 {
			return 0, -1
		}

		total := n + vn
		return uint64(v), total

	case TypeUint32,
		TypeUint64:
		v, vn := rvarint.Uint64(b)
		if vn < 0 {
			return 0, -1
		}

		total := n + vn
		return v, total
	}

	return 0, -1
}

// U128/U256

func DecodeU128(b []byte) (u128.U128, int, error) {
	t, n := decodeType(b)
	switch {
	case n < 0:
		return u128.U128{}, -1, errors.New("decode u128: invalid data")
	case t == TypeNil:
		return u128.U128{}, n, nil
	case t != TypeU128:
		return u128.U128{}, -1, fmt.Errorf("decode u128: unexpected type, expected=%d, actual=%d", TypeU128, t)
	}

	end := len(b) - n
	start := end - 16
	if start < 0 {
		return u128.U128{}, -1, errors.New("decode u128: invalid data")
	}

	p := b[start:end]
	v, err := u128.Parse(p)
	if err != nil {
		return u128.U128{}, -1, err
	}

	total := n + 16
	return v, total, nil
}

func DecodeU256(b []byte) (u256.U256, int, error) {
	t, n := decodeType(b)
	switch {
	case n < 0:
		return u256.U256{}, -1, errors.New("decode u256: invalid data")
	case t == TypeNil:
		return u256.U256{}, -1, nil
	case t != TypeU256:
		return u256.U256{}, -1, fmt.Errorf("decode u256: unexpected type, expected=%d, actual=%d", TypeU256, t)
	}

	end := len(b) - n
	start := end - 32
	if start < 0 {
		return u256.U256{}, -1, fmt.Errorf("decode u256: invalid data")
	}

	p := b[start:end]
	v, err := u256.Parse(p)
	if err != nil {
		return u256.U256{}, -1, err
	}

	total := n + 32
	return v, total, err
}

// Float

func DecodeFloat32(b []byte) (float32, int, error) {
	v, n := decodeFloat64(b)
	switch {
	case n < 0:
		return 0, -1, errors.New("decode float32: invalid data")
	case v < math.SmallestNonzeroFloat32:
		return 0, -1, errors.New("decode float32: overflow, value too small")
	case v > math.MaxFloat64:
		return 0, -1, errors.New("decode float32: overflow, value too large")
	}
	return float32(v), n, nil
}

func DecodeFloat64(b []byte) (float64, int, error) {
	v, n := decodeFloat64(b)
	if n < 0 {
		return 0, n, errors.New("decode float64: invalid data")
	}
	return v, n, nil
}

// decodeFloat64 reads and returns any float as float64 and the number of decode bytes n, or -n on error.
func decodeFloat64(b []byte) (float64, int) {
	// type
	t, n := decodeType(b)
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

// Bytes

func DecodeBytes(b []byte) ([]byte, int, error) {
	// type
	t, n := decodeType(b)
	switch {
	case n < 0:
		return nil, -1, errors.New("decode bytes: invalid data")
	case t == TypeNil:
		return nil, n, nil
	case t != TypeBytes:
		return nil, -1, fmt.Errorf("decode bytes: unexpected type, expected=%d, actual=%d", TypeBytes, t)
	}

	// bytes size
	off := len(b) - n
	size, sn := decodeBytesSize(b[:off])
	if sn < 0 {
		return nil, -1, errors.New("decode bytes: invalid size")
	}

	// bytes body
	off -= sn
	body, err := decodeBytesBody(b[:off], size)
	if err != nil {
		return nil, -1, err
	}

	total := n + sn + int(size)
	return body, total, nil
}

func decodeBytesSize(b []byte) (uint32, int) {
	return rvarint.Uint32(b)
}

func decodeBytesBody(b []byte, size uint32) ([]byte, error) {
	end := len(b)
	start := end - int(size)
	if start < 0 {
		return nil, fmt.Errorf("decode bytes: invalid body, expected size=%d, actual size=%d", size, len(b))
	}

	v := b[start:end]
	return v, nil
}

// String

func DecodeString(b []byte) (string, int, error) {
	// type
	t, n := decodeType(b)
	switch {
	case n < 0:
		return "", -1, errors.New("decode string: invalid data")
	case t == TypeNil:
		return "", n, nil
	case t != TypeString:
		return "", -1, fmt.Errorf("decode string: unexpected type, expected=%d, actual=%d", TypeString, t)
	}

	// string size
	off := len(b) - n
	size, sn := decodeStringSize(b[:off])
	if sn < 0 {
		return "", -1, fmt.Errorf("decode string: invalid size")
	}

	// string body
	off -= (sn + 1) // zero byte
	body, err := decodeStringBody(b[:off], size)
	if err != nil {
		return "", -1, err
	}

	total := n + sn + 1 + int(size)
	return body, total, err
}

func decodeStringSize(b []byte) (uint32, int) {
	return rvarint.Uint32(b)
}

func decodeStringBody(b []byte, size uint32) (string, error) {
	end := len(b)
	start := end - int(size)
	if start < 0 {
		return "", fmt.Errorf("decode string: invalid body, expected size=%d, actual size=%d", size, len(b))
	}

	p := b[start:end]
	s := *(*string)(unsafe.Pointer(&p))
	return s, nil
}

// List meta

func decodeListMeta(b []byte) (listMeta, int, error) {
	meta := listMeta{}
	if len(b) == 0 {
		return meta, 0, nil
	}

	// decode type
	t, n := decodeType(b)
	if n < 0 {
		return meta, -1, errors.New("decode list: invalid data")
	}

	// check type
	switch t {
	default:
		return meta, -1, fmt.Errorf("decode list: unexpected type, expected=%d, actual=%d", TypeList, t)
	case TypeNil:
		return meta, n, nil
	case TypeList, TypeListBig:
	}
	big := t == TypeListBig

	// table size
	off := len(b) - 1
	tsize, tn := decodeListTableSize(b[:off])
	if tn < 0 {
		return meta, -1, errors.New("decode list: invalid table size")
	}

	// body size
	off -= int(tn)
	bsize, dn := decodeListBodySize(b[:off])
	if dn < 0 {
		return meta, -1, errors.New("decode list: invalid body size")
	}

	// table
	off -= int(dn)
	table, err := decodeListTable(b[:off], tsize, big)
	if err != nil {
		return meta, -1, err
	}

	// body
	off -= int(tsize)
	off -= int(bsize)
	if off < 0 {
		return meta, -1, errors.New("decode list: invalid body")
	}

	// done
	meta = listMeta{
		table: table,
		body:  bsize,
		big:   big,
	}

	total := n + tn + dn + int(tsize) + int(bsize)
	return meta, total, nil
}

func decodeListTableSize(b []byte) (uint32, int) {
	return rvarint.Uint32(b)
}

func decodeListBodySize(b []byte) (uint32, int) {
	return rvarint.Uint32(b)
}

func decodeListTable(b []byte, size uint32, big bool) (listTable, error) {
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
		return nil, errors.New("decode list: invalid table, array too small")
	}

	// check divisible
	if size%elemSize != 0 {
		return nil, fmt.Errorf("decode list: invalid table, size not divisible by %d, size=%d",
			elemSize, size)
	}

	p := b[off:]
	v := listTable(p)
	return v, nil
}

// Message meta

func decodeMessageMeta(b []byte) (messageMeta, int, error) {
	meta := messageMeta{}
	if len(b) == 0 {
		return meta, 0, nil
	}

	// decode type
	t, n := decodeType(b)
	if n < 0 {
		return meta, -1, errors.New("decode message: invalid type")
	}

	// check type
	switch t {
	default:
		return meta, -1, fmt.Errorf("decode message: unexpected type, expected=%d, actual=%d", TypeMessage, t)
	case TypeNil:
		return meta, 0, nil
	case TypeMessage, TypeMessageBig:
	}
	big := t == TypeMessageBig

	// table size
	off := len(b) - n
	tsize, tn := decodeMessageTableSize(b[:off])
	if tn < 0 {
		return meta, -1, errors.New("decode message: invalid table size")
	}

	// body size
	off -= int(tn)
	bsize, dn := decodeMessageBodySize(b[:off])
	if dn < 0 {
		return meta, -1, fmt.Errorf("decode message: invalid body size")
	}

	// table
	off -= int(dn)
	table, err := decodeMessageTable(b[:off], tsize, big)
	if err != nil {
		return meta, -1, err
	}

	// body
	off -= int(tsize)
	off -= int(bsize)
	if off < 0 {
		return meta, -1, errors.New("decode message: invalid body")
	}

	// done
	meta = messageMeta{
		table: table,
		body:  bsize,
		big:   big,
	}

	total := n + tn + dn + int(tsize) + int(bsize)
	return meta, total, nil
}

func decodeMessageTableSize(b []byte) (uint32, int) {
	return rvarint.Uint32(b)
}

func decodeMessageBodySize(b []byte) (uint32, int) {
	return rvarint.Uint32(b)
}

func decodeMessageTable(b []byte, size uint32, big bool) (messageTable, error) {
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
		return nil, errors.New("decode message: invalid table, array too small")
	}

	// check divisible
	if size%fieldSize != 0 {
		return nil, fmt.Errorf("decode message: invalid table, size not divisible by %d, size=%d",
			fieldSize, size)
	}

	p := b[off:]
	v := messageTable(p)
	return v, nil
}

// Struct

func decodeStruct(b []byte) (bodySize int, n int, err error) {
	if len(b) == 0 {
		return 0, 0, nil
	}

	// decode type
	t, n := decodeType(b)
	if n < 0 {
		return 0, -1, errors.New("decode struct: invalid type")
	}

	// check type
	switch t {
	default:
		return 0, -1, fmt.Errorf("decode struct: unexpected type, expected=%d, actual=%d", TypeStruct, t)
	case TypeNil:
		return 0, 0, nil
	case TypeStruct:
	}

	// body size
	off := len(b) - n
	bsize, bn := decodeStructBodySize(b[:off])
	if bn < 0 {
		return 0, -1, errors.New("decode struct: invalid body size")
	}

	// done
	total := n + bn + int(bsize)
	return int(bsize), total, nil
}

func decodeStructBodySize(b []byte) (uint32, int) {
	return rvarint.Uint32(b)
}
