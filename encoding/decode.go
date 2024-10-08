// Copyright 2022 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"unsafe"

	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/encoding/compactint"
	"github.com/basecomplextech/spec/internal/core"
)

// DecodeType decodes a value type.
func DecodeType(b []byte) (core.Type, int, error) {
	v, n := decodeType(b)
	if n < 0 {
		return 0, 0, fmt.Errorf("decode type: invalid data")
	}

	size := n
	return core.Type(v), size, nil
}

// DecodeTypeSize decodes a value type and its total size, returns 0, 0 on error.
func DecodeTypeSize(b []byte) (core.Type, int, error) {
	if len(b) == 0 {
		return core.TypeUndefined, 0, nil
	}

	t, n := decodeType(b)
	if n < 0 {
		return 0, 0, fmt.Errorf("decode type: invalid data")
	}

	end := len(b) - n
	v := b[:end]

	switch t {
	case core.TypeTrue, core.TypeFalse:
		return t, n, nil

	case core.TypeByte:
		if len(v) < 1 {
			return 0, 0, fmt.Errorf("decode byte: invalid data")
		}
		return t, n + 1, nil

	// Int

	case core.TypeInt16, core.TypeInt32, core.TypeInt64:
		m := compactint.ReverseSize(v)
		if m <= 0 {
			return 0, 0, fmt.Errorf("decode int: invalid data")
		}
		return t, n + m, nil

	// Uint

	case core.TypeUint16, core.TypeUint32, core.TypeUint64:
		m := compactint.ReverseSize(v)
		if m <= 0 {
			return 0, 0, fmt.Errorf("decode uint: invalid data")
		}
		return t, n + m, nil

	// Float

	case core.TypeFloat32:
		m := 4
		if len(v) < m {
			return 0, 0, fmt.Errorf("decode float32: invalid data")
		}
		return t, n + m, nil

	case core.TypeFloat64:
		m := 8
		if len(v) < m {
			return 0, 0, fmt.Errorf("decode float64: invalid data")
		}
		return t, n + m, nil

	// Bin

	case core.TypeBin64:
		m := 8
		if len(v) < m {
			return 0, 0, fmt.Errorf("decode bin64: invalid data")
		}
		return t, n + m, nil

	case core.TypeBin128:
		m := 16
		if len(v) < m {
			return 0, 0, fmt.Errorf("decode bin128: invalid data")
		}
		return t, n + m, nil

	case core.TypeBin256:
		m := 32
		if len(v) < m {
			return 0, 0, fmt.Errorf("decode bin256: invalid data")
		}
		return t, n + m, nil

	// Bytes/string

	case core.TypeBytes:
		dataSize, m := decodeSize(v)
		if m < 0 {
			return 0, 0, errors.New("decode bytes: invalid data size")
		}
		size := n + m + int(dataSize)
		if len(b) < size {
			return 0, 0, errors.New("decode bytes: invalid data")
		}
		return t, size, nil

	case core.TypeString:
		dataSize, m := decodeSize(v)
		if m < 0 {
			return 0, 0, errors.New("decode string: invalid data size")
		}
		size := n + m + int(dataSize) + 1 // +1 for null terminator
		if len(b) < size {
			return 0, 0, errors.New("decode string: invalid data")
		}
		return t, size, nil // +1 for null terminator

	// List

	case core.TypeList, core.TypeBigList:
		size := n

		// Table size
		tableSize, m := decodeSize(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode list: invalid table size")
		}
		end -= m
		size += m + int(tableSize)

		// Data size
		dataSize, m := decodeSize(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode list: invalid data size")
		}
		end -= m
		size += m + int(dataSize)

		if len(b) < size {
			return 0, 0, errors.New("decode list: invalid data")
		}
		return t, size, nil

	// Message

	case core.TypeMessage, core.TypeBigMessage:
		size := n

		// Table size
		tableSize, m := decodeSize(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode message: invalid table size")
		}
		end -= m
		size += m + int(tableSize)

		// Data size
		dataSize, m := decodeSize(b[:end])
		if m < 0 {
			return 0, 0, fmt.Errorf("decode message: invalid data size")
		}
		end -= m
		size += m + int(dataSize)

		if len(b) < size {
			return 0, 0, errors.New("decode message: invalid data")
		}
		return t, size, nil

	// Struct

	case core.TypeStruct:
		size := n

		// Data size
		dataSize, m := decodeSize(b[:end])
		if n < 0 {
			return 0, 0, errors.New("decode struct: invalid data size")
		}

		size += m + int(dataSize)
		if len(b) < size {
			return 0, 0, errors.New("decode struct: invalid data")
		}
		return t, size, nil
	}

	return 0, 0, fmt.Errorf("decode: invalid type, type=%d", t)
}

func decodeType(b []byte) (core.Type, int) {
	if len(b) == 0 {
		return core.TypeUndefined, 0
	}

	v := b[len(b)-1]
	return core.Type(v), 1
}

// Byte

func DecodeByte(b []byte) (byte, int, error) {
	if len(b) == 0 {
		return 0, 0, nil
	}

	typ, n := decodeType(b)
	if n < 0 {
		return 0, 0, errors.New("decode byte: invalid data")
	}
	if typ != core.TypeByte {
		return 0, 0, fmt.Errorf("decode byte: invalid type, type=%v:%d", typ, typ)
	}

	end := len(b) - 2
	if end < 0 {
		return 0, 0, errors.New("decode byte: invalid data")
	}
	return b[end], 2, nil
}

// Bool

func DecodeBool(b []byte) (bool, int, error) {
	if len(b) == 0 {
		return false, 0, nil
	}

	typ, n := decodeType(b)
	if n < 0 {
		return false, 0, errors.New("decode bool: invalid data")
	}

	v := typ == core.TypeTrue
	size := n
	return v, size, nil
}

// Int

func DecodeInt16(b []byte) (int16, int, error) {
	if len(b) == 0 {
		return 0, 0, nil
	}

	typ, n := decodeType(b)
	if n < 0 {
		return 0, 0, errors.New("decode int16: invalid data")
	}
	end := len(b) - n

	switch typ {
	case core.TypeInt16, core.TypeInt32:
		v, m := compactint.ReverseInt32(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode int16: invalid data")
		}

		switch {
		case v < math.MinInt16:
			return 0, 0, errors.New("decode int16: overflow, value too small")
		case v > math.MaxInt16:
			return 0, 0, errors.New("decode int16: overflow, value too large")
		}

		n += m
		return int16(v), n, nil

	case core.TypeInt64:
		v, m := compactint.ReverseInt64(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode int16: invalid data")
		}

		switch {
		case v < math.MinInt16:
			return 0, 0, errors.New("decode int16: overflow, value too small")
		case v > math.MaxInt16:
			return 0, 0, errors.New("decode int16: overflow, value too large")
		}

		n += m
		return int16(v), n, nil
	}

	return 0, 0, fmt.Errorf("decode int16: invalid type, type=%v:%d", typ, typ)
}

func DecodeInt32(b []byte) (int32, int, error) {
	if len(b) == 0 {
		return 0, 0, nil
	}

	typ, n := decodeType(b)
	if n < 0 {
		return 0, 0, errors.New("decode int32: invalid data")
	}
	end := len(b) - n

	switch typ {
	case core.TypeInt16, core.TypeInt32:
		v, m := compactint.ReverseInt32(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode int32: invalid data")
		}

		n += m
		return v, n, nil

	case core.TypeInt64:
		v, m := compactint.ReverseInt64(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode int32: invalid data")
		}

		switch {
		case v < math.MinInt32:
			return 0, 0, errors.New("decode int32: overflow, value too small")
		case v > math.MaxInt32:
			return 0, 0, errors.New("decode int32: overflow, value too large")
		}

		n += m
		return int32(v), n, nil
	}

	return 0, 0, fmt.Errorf("decode int32: invalid type, type=%v:%d", typ, typ)
}

func DecodeInt64(b []byte) (int64, int, error) {
	if len(b) == 0 {
		return 0, 0, nil
	}

	typ, n := decodeType(b)
	if n < 0 {
		return 0, 0, errors.New("decode int64: invalid data")
	}
	end := len(b) - n

	switch typ {
	case core.TypeInt16, core.TypeInt32:
		v, m := compactint.ReverseInt32(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode int64: invalid data")
		}
		n += m
		return int64(v), n, nil

	case core.TypeInt64:
		v, m := compactint.ReverseInt64(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode int64: invalid data")
		}
		n += m
		return int64(v), n, nil
	}

	return 0, 0, fmt.Errorf("decode int64: invalid type, type=%v:%d", typ, typ)
}

// Uint

func DecodeUint16(b []byte) (uint16, int, error) {
	if len(b) == 0 {
		return 0, 0, nil
	}

	typ, n := decodeType(b)
	if n < 0 {
		return 0, 0, errors.New("decode uint16: invalid data")
	}
	end := len(b) - n

	switch typ {
	case core.TypeUint16, core.TypeUint32:
		v, m := compactint.ReverseUint32(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode uint16: invalid data")
		}

		if v > math.MaxUint16 {
			return 0, 0, errors.New("decode int16: overflow, value too large")
		}

		n += m
		return uint16(v), n, nil

	case core.TypeUint64:
		v, m := compactint.ReverseUint64(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode uint16: invalid data")
		}

		if v > math.MaxUint16 {
			return 0, 0, errors.New("decode int16: overflow, value too large")
		}

		n += m
		return uint16(v), n, nil
	}

	return 0, 0, fmt.Errorf("decode uint32: invalid type, type=%v:%d", typ, typ)
}

func DecodeUint32(b []byte) (uint32, int, error) {
	if len(b) == 0 {
		return 0, 0, nil
	}

	typ, n := decodeType(b)
	if n < 0 {
		return 0, 0, errors.New("decode uint32: invalid data")
	}
	end := len(b) - n

	switch typ {
	case core.TypeUint16, core.TypeUint32:
		v, m := compactint.ReverseUint32(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode uint32: invalid data")
		}
		n += m
		return v, n, nil

	case core.TypeUint64:
		v, m := compactint.ReverseUint64(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode uint32: invalid data")
		}

		if v > math.MaxUint32 {
			return 0, 0, errors.New("decode int32: overflow, value too large")
		}

		n += m
		return uint32(v), n, nil
	}

	return 0, 0, fmt.Errorf("decode uint32: invalid type, type=%v:%d", typ, typ)
}

func DecodeUint64(b []byte) (uint64, int, error) {
	if len(b) == 0 {
		return 0, 0, nil
	}

	typ, n := decodeType(b)
	if n < 0 {
		return 0, 0, errors.New("decode uint64: invalid data")
	}
	end := len(b) - n

	switch typ {
	case core.TypeUint16, core.TypeUint32:
		v, m := compactint.ReverseUint32(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode uint64: invalid data")
		}
		n += m
		return uint64(v), n, nil

	case core.TypeUint64:
		v, m := compactint.ReverseUint64(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode uint64: invalid data")
		}
		n += m
		return v, n, nil
	}

	return 0, 0, fmt.Errorf("decode uint64: invalid type, type=%v:%d", typ, typ)
}

// Bin64/128/256

func DecodeBin64(b []byte) (_ bin.Bin64, size int, err error) {
	if len(b) == 0 {
		return bin.Bin64{}, 0, nil
	}

	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode bin64: invalid data")
		return
	}
	if typ != core.TypeBin64 {
		err = fmt.Errorf("decode bin64: invalid type, type=%v:%d", typ, typ)
		return
	}

	size = n
	start := len(b) - (n + 8)
	end := len(b) - n

	if start < 0 {
		err = errors.New("decode bin64: invalid data")
		return
	}

	v, err := bin.Parse64(b[start:end])
	if err != nil {
		return
	}

	size += 8
	return v, size, nil
}

func DecodeBin128(b []byte) (_ bin.Bin128, size int, err error) {
	if len(b) == 0 {
		return bin.Bin128{}, 0, nil
	}

	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode bin128: invalid data")
		return
	}
	if typ != core.TypeBin128 {
		err = fmt.Errorf("decode bin128: invalid type, type=%v:%d", typ, typ)
		return
	}

	size = n
	start := len(b) - (n + 16)
	end := len(b) - n

	if start < 0 {
		err = errors.New("decode bin128: invalid data")
		return
	}

	v, err := bin.Parse128(b[start:end])
	if err != nil {
		return
	}

	size += 16
	return v, size, nil
}

func DecodeBin256(b []byte) (_ bin.Bin256, size int, err error) {
	if len(b) == 0 {
		return bin.Bin256{}, 0, nil
	}

	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode bin256: invalid data")
		return
	}
	if typ != core.TypeBin256 {
		err = fmt.Errorf("decode bin256: invalid type, type=%v:%d", typ, typ)
		return
	}

	size = n
	start := len(b) - (n + 32)
	end := len(b) - n

	if start < 0 {
		err = fmt.Errorf("decode bin256: invalid data")
		return
	}

	v, err := bin.Parse256(b[start:end])
	if err != nil {
		return
	}

	size += 32
	return v, size, nil
}

// Float

func DecodeFloat32(b []byte) (float32, int, error) {
	if len(b) == 0 {
		return 0, 0, nil
	}

	v, n := decodeFloat64(b)
	switch {
	case n < 0:
		return 0, 0, errors.New("decode float32: invalid data")
	case v < -math.MaxFloat32:
		return 0, 0, errors.New("decode float32: overflow, value too small")
	case v > math.MaxFloat32:
		return 0, 0, errors.New("decode float32: overflow, value too large")
	}

	size := n
	return float32(v), size, nil
}

func DecodeFloat64(b []byte) (float64, int, error) {
	if len(b) == 0 {
		return 0, 0, nil
	}

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
	case core.TypeFloat32:
		start := len(b) - 5
		if start < 0 {
			return 0, -1
		}

		v := binary.BigEndian.Uint32(b[start:])
		f := math.Float32frombits(v)
		return float64(f), 5

	case core.TypeFloat64:
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

func DecodeBytes(b []byte) (_ core.Bytes, size int, err error) {
	if len(b) == 0 {
		return nil, 0, nil
	}

	// core.Type
	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode bytes: invalid data")
		return
	}
	if typ != core.TypeBytes {
		err = fmt.Errorf("decode bytes: invalid type, type=%v:%d", typ, typ)
		return
	}

	size = n
	end := len(b) - size

	// Data size
	dataSize, n := decodeSize(b[:end])
	if n < 0 {
		err = errors.New("decode bytes: invalid data size")
		return
	}
	size += n
	end -= n

	// Data
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

func DecodeString(b []byte) (_ core.String, size int, err error) {
	if len(b) == 0 {
		return "", 0, nil
	}

	// core.Type
	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode string: invalid data")
		return
	}
	if typ != core.TypeString {
		err = fmt.Errorf("decode string: invalid type, type=%v:%d", typ, typ)
		return
	}

	size = n
	end := len(b) - size

	// Size
	dataSize, n := decodeSize(b[:end])
	if n < 0 {
		err = fmt.Errorf("decode string: invalid data size")
		return
	}
	size += n + 1
	end -= (n + 1) // null terminator

	// Data
	data, err := decodeStringData(b[:end], dataSize)
	if err != nil {
		return
	}

	size += int(dataSize)
	return core.String(data), size, nil
}

func DecodeStringClone(b []byte) (_ string, size int, err error) {
	s, size, err := DecodeString(b)
	if err != nil {
		return "", size, err
	}
	return s.Clone(), size, nil
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

	// Decode type
	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode struct: invalid type")
		return
	}
	if typ != core.TypeStruct {
		err = fmt.Errorf("decode struct: invalid type, type=%v:%d", typ, typ)
		return
	}

	size = n
	end := len(b) - size

	// Data size
	dataSize_, n := decodeSize(b[:end])
	if n < 0 {
		err = errors.New("decode struct: invalid data size")
		return
	}
	size += n + int(dataSize_)

	return int(dataSize_), size, nil
}

// ListMeta

func DecodeListMeta(b []byte) (_ ListMeta, size int, err error) {
	if len(b) == 0 {
		return
	}

	// Decode type
	typ, n := decodeType(b)
	if n < 0 {
		n = 0
		err = errors.New("decode list: invalid data")
		return
	}
	if typ != core.TypeList && typ != core.TypeBigList {
		err = fmt.Errorf("decode list: invalid type, type=%v:%d", typ, typ)
		return
	}

	// Start
	size = n
	end := len(b) - n
	big := typ == core.TypeBigList

	// Table size
	tableSize, n := decodeSize(b[:end])
	if n < 0 {
		err = errors.New("decode list: invalid table size")
		return
	}
	end -= n
	size += n

	// Data size
	dataSize, n := decodeSize(b[:end])
	if n < 0 {
		err = errors.New("decode list: invalid data size")
		return
	}
	end -= n
	size += n

	// Table
	table, err := decodeListTable(b[:end], tableSize, big)
	if err != nil {
		return
	}
	end -= int(tableSize) + int(dataSize)
	size += int(tableSize)

	// Data
	if end < 0 {
		err = errors.New("decode list: invalid data")
		return
	}
	size += int(dataSize)

	// Done
	meta := ListMeta{
		table: table,
		data:  dataSize,
		big:   big,
	}
	return meta, size, nil
}

func decodeListTable(b []byte, size uint32, big bool) (_ listTable, err error) {
	// Element size
	elemSize := listElementSmallSize
	if big {
		elemSize = listElementBigSize
	}

	// Check offset
	start := len(b) - int(size)
	if start < 0 {
		err = errors.New("decode list: invalid table")
		return
	}

	// Check divisible
	if size%uint32(elemSize) != 0 {
		err = errors.New("decode list: invalid table")
		return
	}

	p := b[start:]
	v := listTable(p)
	return v, nil
}

// MessageMeta

func DecodeMessageMeta(b []byte) (_ MessageMeta, size int, err error) {
	if len(b) == 0 {
		return
	}

	// Decode type
	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode message: invalid type")
		return
	}
	switch typ {
	case core.TypeMessage, core.TypeBigMessage:
	default:
		err = fmt.Errorf("decode message: invalid type, type=%v:%d", typ, typ)
		return
	}

	// Start
	size = n
	end := len(b) - size
	big := typ == core.TypeBigMessage

	// Table size
	tableSize, m := decodeSize(b[:end])
	if m < 0 {
		err = errors.New("decode message: invalid table size")
		return
	}
	end -= m
	size += m

	// Data size
	dataSize, m := decodeSize(b[:end])
	if m < 0 {
		err = fmt.Errorf("decode message: invalid data size")
		return
	}
	end -= m
	size += m

	// Table
	table, err := decodeMessageTable(b[:end], tableSize, big)
	if err != nil {
		return
	}
	end -= int(tableSize) + int(dataSize)
	size += int(tableSize)

	// Data
	if end < 0 {
		err = errors.New("decode message: invalid data")
		return
	}
	size += int(dataSize)

	// Done
	meta := MessageMeta{
		table: table,
		data:  dataSize,
		big:   big,
	}
	return meta, size, nil
}

func decodeMessageTable(b []byte, size uint32, big bool) (_ messageTable, err error) {
	// Field size
	fieldSize := messageFieldSmallSize
	if big {
		fieldSize = messageFieldBigSize
	}

	// Check offset
	start := len(b) - int(size)
	if start < 0 {
		err = errors.New("decode message: invalid table")
		return
	}

	// Check divisible
	if size%uint32(fieldSize) != 0 {
		err = errors.New("decode message: invalid table")
		return
	}

	p := b[start:]
	v := messageTable(p)
	return v, nil
}

// private

func decodeSize(b []byte) (uint32, int) {
	return compactint.ReverseUint32(b)
}
