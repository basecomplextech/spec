package spec

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"unsafe"

	"github.com/baseblck/library/rvarint"
	"github.com/baseblck/library/u128"
	"github.com/baseblck/library/u256"
)

func DecodeType(b []byte) (Type, int, error) {
	v, n := decodeType(b)
	if n < 0 {
		return 0, 0, fmt.Errorf("decode type: invalid data")
	}

	size := n
	return Type(v), size, nil
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
	v, ok := decodeInt64(b)
	switch {
	case ok < 0:
		return 0, 0, errors.New("decode byte: invalid data")
	case v < math.MinInt8:
		return 0, 0, errors.New("decode byte: overflow, value too small")
	case v > math.MaxInt8:
		return 0, 0, errors.New("decode byte: overflow, value too large")
	}

	size := ok
	return byte(v), size, nil
}

// Bool

func DecodeBool(b []byte) (bool, int, error) {
	typ, n := decodeType(b)
	if n < 0 {
		return false, 0, errors.New("decode bool: invalid data")
	}

	v := typ == TypeTrue
	size := n
	return v, size, nil
}

// Int

func DecodeInt32(b []byte) (int32, int, error) {
	v, n := decodeInt64(b)
	switch {
	case n < 0:
		return 0, 0, errors.New("decode int32: invalid data")
	case v < math.MinInt32:
		return 0, 0, errors.New("decode int32: overflow, value too small")
	case v > math.MaxInt32:
		return 0, 0, errors.New("decode int32: overflow, value too large")
	}

	size := n
	return int32(v), size, nil
}

func DecodeInt64(b []byte) (int64, int, error) {
	v, n := decodeInt64(b)
	if n < 0 {
		return 0, 0, errors.New("decode int32: invalid data")
	}

	size := n
	return v, size, nil
}

// decodeInt64 reads and returns any int as int64 and the number of decode bytes n, or -1 on error.
func decodeInt64(b []byte) (int64, int) {
	// type
	typ, n := decodeType(b)
	if n < 0 {
		return 0, -1
	}

	end := len(b) - n

	// read, cast int
	switch typ {
	case TypeNil:
		return 0, n

	case TypeTrue:
		return 1, n

	case TypeFalse:
		return 0, n

	case TypeByte:
		if len(b) < 1 {
			return 0, -1
		}

		v := b[end-1]
		n += 1
		return int64(v), n

	case TypeInt32, TypeInt64:
		v, m := rvarint.Int64(b[:end])
		if m < 0 {
			return 0, -1
		}

		n += m
		return v, n

	case TypeUint32, TypeUint64:
		v, m := rvarint.Uint64(b[:end])
		if m < 0 {
			return 0, -1
		}

		n += m
		return int64(v), n
	}

	return 0, -1
}

// Uint

func DecodeUint32(b []byte) (uint32, int, error) {
	v, n := decodeUint64(b)
	switch {
	case n < 0:
		return 0, 0, errors.New("decode uint32: invalid data")
	case v > math.MaxUint32:
		return 0, 0, errors.New("decode uint32: overflow, value too large")
	}

	size := n
	return uint32(v), size, nil
}

func DecodeUint64(b []byte) (uint64, int, error) {
	v, n := decodeUint64(b)
	if n < 0 {
		return 0, 0, fmt.Errorf("decode uint64: invalid data")
	}

	size := n
	return v, size, nil
}

// decodeUint64 reads and returns any int as uint64 and the number of decode bytes n, or -n on error.
func decodeUint64(b []byte) (uint64, int) {
	// type
	typ, n := decodeType(b)
	if n < 0 {
		return 0, -1
	}

	end := len(b) - n

	switch typ {
	case TypeNil:
		return 0, n

	case TypeTrue:
		return 1, n

	case TypeFalse:
		return 0, n

	case TypeByte:
		if len(b) < 1 {
			return 0, -1
		}

		v := b[end-1]
		n += 1
		return uint64(v), n

	case TypeInt32, TypeInt64:
		v, m := rvarint.Int64(b[:end])
		if m < 0 {
			return 0, -1
		}

		n += m
		return uint64(v), n

	case TypeUint32, TypeUint64:
		v, m := rvarint.Uint64(b[:end])
		if m < 0 {
			return 0, -1
		}

		n += m
		return v, n
	}

	return 0, -1
}

// U128/U256

func DecodeU128(b []byte) (_ u128.U128, size int, err error) {
	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode u128: invalid data")
		return
	}

	switch {
	case typ == TypeNil:
		return
	case typ != TypeU128:
		err = fmt.Errorf("decode u128: invalid type, type=%v:%d", typ, typ)
		return
	}

	size = n
	start := len(b) - (n + 16)
	end := len(b) - n

	if start < 0 {
		err = errors.New("decode u128: invalid data")
		return
	}

	v, err := u128.Parse(b[start:end])
	if err != nil {
		return
	}

	size += 16
	return v, size, nil
}

func DecodeU256(b []byte) (_ u256.U256, size int, err error) {
	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode u256: invalid data")
		return
	}

	switch {
	case typ == TypeNil:
		return
	case typ != TypeU256:
		err = fmt.Errorf("decode u256: invalid type, type=%v:%d", typ, typ)
		return
	}

	size = n
	start := len(b) - (n + 32)
	end := len(b) - n

	if start < 0 {
		err = fmt.Errorf("decode u256: invalid data")
		return
	}

	v, err := u256.Parse(b[start:end])
	if err != nil {
		return
	}

	size += 32
	return v, size, nil
}

// Float

func DecodeFloat32(b []byte) (float32, int, error) {
	v, n := decodeFloat64(b)
	switch {
	case n < 0:
		return 0, 0, errors.New("decode float32: invalid data")
	case v < math.SmallestNonzeroFloat32:
		return 0, 0, errors.New("decode float32: overflow, value too small")
	case v > math.MaxFloat64:
		return 0, 0, errors.New("decode float32: overflow, value too large")
	}

	size := n
	return float32(v), size, nil
}

func DecodeFloat64(b []byte) (float64, int, error) {
	v, n := decodeFloat64(b)
	if n < 0 {
		return 0, n, errors.New("decode float64: invalid data")
	}

	size := n
	return v, size, nil
}

// decodeFloat64 reads and returns any float as float64 and the number of decode bytes n, or -n on error.
func decodeFloat64(b []byte) (float64, int) {
	t, n := decodeType(b)
	if n < 0 {
		return 0, n
	}

	switch t {
	case TypeNil:
		return 0, n

	case TypeFloat32:
		start := len(b) - 5
		if start < 0 {
			return 0, -1
		}

		v := binary.BigEndian.Uint32(b[start:])
		f := math.Float32frombits(v)
		return float64(f), 5

	case TypeFloat64:
		start := len(b) - 9
		if start < 0 {
			return 0, -1
		}

		v := binary.BigEndian.Uint64(b[start:])
		f := math.Float64frombits(v)
		return f, 9
	}

	return 0, -1
}

// Bytes

func DecodeBytes(b []byte) (_ []byte, size int, err error) {
	// type
	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode bytes: invalid data")
		return
	}

	switch typ {
	default:
		err = fmt.Errorf("decode bytes: invalid type, type=%v:%d", typ, typ)
		return
	case TypeNil:
		return
	case TypeBytes, TypeBytesBig:
	}

	size = n
	end := len(b) - size
	big := typ == TypeBytesBig

	// data size
	dataSize, n := decodeSize(b[:end], big)
	if n < 0 {
		err = errors.New("decode bytes: invalid data size")
		return
	}
	size += n
	end -= n

	// data
	data, err := decodeBytesData(b[:end], dataSize)
	if err != nil {
		return nil, 0, err
	}

	size += int(dataSize)
	return data, size, nil
}

func decodeBytesData(b []byte, size uint32) ([]byte, error) {
	off := len(b) - int(size)
	if off < 0 {
		return nil, errors.New("decode bytes: invalid data size")
	}
	return b[off:], nil
}

// String

func DecodeString(b []byte) (_ string, size int, err error) {
	// type
	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode string: invalid data")
		return
	}

	switch typ {
	default:
		err = fmt.Errorf("decode string: invalid type, type=%v:%d", typ, typ)
		return
	case TypeNil:
		return
	case TypeString, TypeBigString:
	}

	size = n
	end := len(b) - size
	big := typ == TypeBigString

	// size
	dataSize, n := decodeSize(b[:end], big)
	if n < 0 {
		err = fmt.Errorf("decode string: invalid data size")
		return
	}
	size += n + 1
	end -= (n + 1) // zero byte

	// data
	data, err := decodeStringData(b[:end], dataSize)
	if err != nil {
		return
	}

	size += int(dataSize)
	return data, size, nil
}

func decodeStringData(b []byte, size uint32) (string, error) {
	off := len(b) - int(size)
	if off < 0 {
		return "", errors.New("decode string: invalid data size")
	}

	p := b[off:]
	s := *(*string)(unsafe.Pointer(&p))
	return s, nil
}

// Struct

func DecodeStruct(b []byte) (dataSize int, size int, err error) {
	if len(b) == 0 {
		return 0, 0, nil
	}

	// decode type
	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode struct: invalid type")
		return
	}

	// check type
	switch typ {
	default:
		err = fmt.Errorf("decode struct: invalid type, type=%v:%d", typ, typ)
		return
	case TypeNil:
		return
	case TypeStruct, TypeBigStruct:
	}

	size = n
	end := len(b) - size
	big := typ == TypeBigStruct

	// data size
	dsize, n := decodeSize(b[:end], big)
	if n < 0 {
		err = errors.New("decode struct: invalid data size")
		return
	}
	size += n + int(dsize)

	return int(dsize), size, nil
}

// list meta

func decodeListMeta(b []byte) (_ listMeta, size int, err error) {
	if len(b) == 0 {
		return
	}

	// decode type
	typ, n := decodeType(b)
	if n < 0 {
		n = 0
		err = errors.New("decode list: invalid data")
		return
	}

	// check type
	switch typ {
	default:
		err = fmt.Errorf("decode list: invalid type, type=%v:%d", typ, typ)
		return
	case TypeNil:
		return
	case TypeList, TypeBigList:
	}

	// start
	size = n
	end := len(b) - n
	big := typ == TypeBigList

	// table size
	tableSize, n := decodeSize(b[:end], big)
	if n < 0 {
		err = errors.New("decode list: invalid table size")
		return
	}
	end -= n
	size += n

	// data size
	dataSize, n := decodeSize(b[:end], big)
	if n < 0 {
		err = errors.New("decode list: invalid data size")
		return
	}
	end -= n
	size += n

	// table
	table, err := decodeListTable(b[:end], tableSize, big)
	if err != nil {
		return
	}
	end -= int(tableSize) + int(dataSize)
	size += int(tableSize)

	// data
	if end < 0 {
		err = errors.New("decode list: invalid data")
		return
	}
	size += int(dataSize)

	// done
	meta := listMeta{
		table: table,
		data:  dataSize,
		big:   big,
	}
	return meta, size, nil
}

func decodeListTable(b []byte, size uint32, big bool) (_ listTable, err error) {
	// element size
	elemSize := listElementSmallSize
	if big {
		elemSize = listElementBigSize
	}

	// check offset
	start := len(b) - int(size)
	if start < 0 {
		err = errors.New("decode list: invalid table")
		return
	}

	// check divisible
	if size%uint32(elemSize) != 0 {
		err = errors.New("decode list: invalid table")
		return
	}

	p := b[start:]
	v := listTable(p)
	return v, nil
}

// message meta

func decodeMessageMeta(b []byte) (_ messageMeta, size int, err error) {
	if len(b) == 0 {
		return
	}

	// decode type
	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode message: invalid type")
		return
	}

	// check type
	switch typ {
	default:
		err = fmt.Errorf("decode message: invalid type, type=%v:%d", typ, typ)
		return
	case TypeNil:
		return
	case TypeMessage, TypeBigMessage:
	}

	// start
	size = n
	end := len(b) - size
	big := typ == TypeBigMessage

	// table size
	tableSize, m := decodeSize(b[:end], big)
	if m < 0 {
		err = errors.New("decode message: invalid table size")
		return
	}
	end -= m
	size += m

	// data size
	dataSize, m := decodeSize(b[:end], big)
	if m < 0 {
		err = fmt.Errorf("decode message: invalid data size")
		return
	}
	end -= m
	size += m

	// table
	table, err := decodeMessageTable(b[:end], tableSize, big)
	if err != nil {
		return
	}
	end -= int(tableSize) + int(dataSize)
	size += int(tableSize)

	// data
	if end < 0 {
		err = errors.New("decode message: invalid data")
		return
	}
	size += int(dataSize)

	// done
	meta := messageMeta{
		table: table,
		data:  dataSize,
		big:   big,
	}
	return meta, size, nil
}

func decodeMessageTable(b []byte, size uint32, big bool) (_ messageTable, err error) {
	// field size
	fieldSize := messageFieldSmallSize
	if big {
		fieldSize = messageFieldBigSize
	}

	// check offset
	start := len(b) - int(size)
	if start < 0 {
		err = errors.New("decode message: invalid table")
		return
	}

	// check divisible
	if size%uint32(fieldSize) != 0 {
		err = errors.New("decode message: invalid table")
		return
	}

	p := b[start:]
	v := messageTable(p)
	return v, nil
}

// private

func decodeSize(b []byte, big bool) (uint32, int) {
	if len(b) < 0 {
		return 0, -1
	}

	if big {
		start := len(b) - 4
		if start < 0 {
			return 0, -1
		}

		size := binary.BigEndian.Uint32(b[start:])
		return size, 4
	}

	start := len(b) - 2
	if start < 0 {
		return 0, -1
	}

	size := binary.BigEndian.Uint16(b[start:])
	return uint32(size), 2
}
