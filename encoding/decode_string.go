// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/basecomplextech/spec/internal/format"
)

func DecodeString(b []byte) (_ format.String, size int, err error) {
	if len(b) == 0 {
		return "", 0, nil
	}

	// format.Type
	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode string: invalid data")
		return
	}
	if typ != format.TypeString {
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
	return format.String(data), size, nil
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
