// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encode

import (
	"encoding/binary"
	"math"

	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/format"
)

func EncodeFloat32(b buffer.Buffer, v float32) (int, error) {
	p := b.Grow(5)
	binary.BigEndian.PutUint32(p, math.Float32bits(v))
	p[4] = byte(format.TypeFloat32)
	return 5, nil
}

func EncodeFloat64(b buffer.Buffer, v float64) (int, error) {
	p := b.Grow(9)
	binary.BigEndian.PutUint64(p, math.Float64bits(v))
	p[8] = byte(format.TypeFloat64)
	return 9, nil
}
