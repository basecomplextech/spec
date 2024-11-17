// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"errors"
	"fmt"

	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/spec/internal/core"
)

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
