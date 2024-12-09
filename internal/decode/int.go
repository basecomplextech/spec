// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package decode

import (
	"errors"
	"fmt"
	"math"

	"github.com/basecomplextech/baselibrary/encoding/compactint"
	"github.com/basecomplextech/spec/internal/format"
)

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
	case format.TypeInt16, format.TypeInt32:
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

	case format.TypeInt64:
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

	return 0, 0, fmt.Errorf("decode int16: invalid type, type=%v", typ)
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
	case format.TypeInt16, format.TypeInt32:
		v, m := compactint.ReverseInt32(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode int32: invalid data")
		}

		n += m
		return v, n, nil

	case format.TypeInt64:
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

	return 0, 0, fmt.Errorf("decode int32: invalid type, type=%v", typ)
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
	case format.TypeInt16, format.TypeInt32:
		v, m := compactint.ReverseInt32(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode int64: invalid data")
		}
		n += m
		return int64(v), n, nil

	case format.TypeInt64:
		v, m := compactint.ReverseInt64(b[:end])
		if m < 0 {
			return 0, 0, errors.New("decode int64: invalid data")
		}
		n += m
		return int64(v), n, nil
	}

	return 0, 0, fmt.Errorf("decode int64: invalid type, type=%v", typ)
}
