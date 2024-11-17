// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/baselibrary/encoding/compactint"
	"github.com/basecomplextech/spec/internal/format"
)

func EncodeInt16(b buffer.Buffer, v int16) (int, error) {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseInt32(p[:], int32(v))
	off := compactint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(format.TypeInt16)

	return n + 1, nil
}

func EncodeInt32(b buffer.Buffer, v int32) (int, error) {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseInt32(p[:], v)
	off := compactint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(format.TypeInt32)

	return n + 1, nil
}

func EncodeInt64(b buffer.Buffer, v int64) (int, error) {
	p := [compactint.MaxLen64]byte{}
	n := compactint.PutReverseInt64(p[:], v)
	off := compactint.MaxLen64 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(format.TypeInt64)

	return n + 1, nil
}
