// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"errors"
	"fmt"
	"math"

	"github.com/basecomplextech/baselibrary/encoding/compactint"
	"github.com/basecomplextech/spec/internal/core"
)

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
