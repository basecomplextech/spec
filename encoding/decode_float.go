// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/basecomplextech/spec/internal/format"
)

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
	case format.TypeFloat32:
		start := len(b) - 5
		if start < 0 {
			return 0, -1
		}

		v := binary.BigEndian.Uint32(b[start:])
		f := math.Float32frombits(v)
		return float64(f), 5

	case format.TypeFloat64:
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
