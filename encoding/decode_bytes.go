// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"errors"
	"fmt"

	"github.com/basecomplextech/spec/internal/core"
)

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
