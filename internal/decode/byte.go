// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package decode

import (
	"errors"
	"fmt"

	"github.com/basecomplextech/spec/internal/format"
)

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
		return 0, 0, fmt.Errorf("decode byte: invalid type, type=%v", typ)
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
