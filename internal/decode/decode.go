// Copyright 2022 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package decode

import (
	"errors"
	"fmt"

	"github.com/basecomplextech/baselibrary/encoding/compactint"
	"github.com/basecomplextech/spec/internal/format"
)

// DecodeType decodes a value type.
func DecodeType(b []byte) (format.Type, int, error) {
	v, n := decodeType(b)
	if n < 0 {
		return 0, 0, fmt.Errorf("decode type: invalid data")
	}

	size := n
	return format.Type(v), size, nil
}

// DecodeTypeSize decodes a value type and its total size, returns 0, 0 on error.
func DecodeTypeSize(b []byte) (format.Type, int, error) {
	if len(b) == 0 {
		return format.TypeUndefined, 0, nil
	}

	t, n := decodeType(b)
	if n < 0 {
		return 0, 0, fmt.Errorf("decode type: invalid data")
	}

	end := len(b) - n
	v := b[:end]

	switch t {
	case format.TypeTrue, format.TypeFalse:
		return t, n, nil

	case format.TypeByte:
		if len(v) < 1 {
			return 0, 0, fmt.Errorf("decode byte: invalid data")
		}
		return t, n + 1, nil

	// Int

	case format.TypeInt16, format.TypeInt32, format.TypeInt64:
		m := compactint.ReverseSize(v)
		if m <= 0 {
			return 0, 0, fmt.Errorf("decode int: invalid data")
		}
		return t, n + m, nil

	// Uint

	case format.TypeUint16, format.TypeUint32, format.TypeUint64:
		m := compactint.ReverseSize(v)
		if m <= 0 {
			return 0, 0, fmt.Errorf("decode uint: invalid data")
		}
		return t, n + m, nil

	// Float

	case format.TypeFloat32:
		m := 4
		if len(v) < m {
			return 0, 0, fmt.Errorf("decode float32: invalid data")
		}
		return t, n + m, nil

	case format.TypeFloat64:
		m := 8
		if len(v) < m {
			return 0, 0, fmt.Errorf("decode float64: invalid data")
		}
		return t, n + m, nil

	// Bin

	case format.TypeBin64:
		m := 8
		if len(v) < m {
			return 0, 0, fmt.Errorf("decode bin64: invalid data")
		}
		return t, n + m, nil

	case format.TypeBin128:
		m := 16
		if len(v) < m {
			return 0, 0, fmt.Errorf("decode bin128: invalid data")
		}
		return t, n + m, nil

	case format.TypeBin256:
		m := 32
		if len(v) < m {
			return 0, 0, fmt.Errorf("decode bin256: invalid data")
		}
		return t, n + m, nil

	// Bytes/string

	case format.TypeBytes:
		dataSize, m := decodeSize(v)
		if m < 0 {
			return 0, 0, errors.New("decode bytes: invalid data size")
		}
		size := n + m + int(dataSize)
		if len(b) < size {
			return 0, 0, errors.New("decode bytes: invalid data")
		}
		return t, size, nil

	case format.TypeString:
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

	case format.TypeList, format.TypeBigList:
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

	case format.TypeMessage, format.TypeBigMessage:
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

	case format.TypeStruct:
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

func decodeType(b []byte) (format.Type, int) {
	if len(b) == 0 {
		return format.TypeUndefined, 0
	}

	v := b[len(b)-1]
	return format.Type(v), 1
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
	if typ != format.TypeByte {
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

	v := typ == format.TypeTrue
	size := n
	return v, size, nil
}

// private

func decodeSize(b []byte) (uint32, int) {
	return compactint.ReverseUint32(b)
}
