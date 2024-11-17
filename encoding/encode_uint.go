// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/baselibrary/encoding/compactint"
	"github.com/basecomplextech/spec/internal/format"
)

func EncodeUint16(b buffer.Buffer, v uint16) (int, error) {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], uint32(v))
	off := compactint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(format.TypeUint16)

	return n + 1, nil
}

func EncodeUint32(b buffer.Buffer, v uint32) (int, error) {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], v)
	off := compactint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(format.TypeUint32)

	return n + 1, nil
}

func EncodeUint64(b buffer.Buffer, v uint64) (int, error) {
	p := [compactint.MaxLen64]byte{}
	n := compactint.PutReverseUint64(p[:], v)
	off := compactint.MaxLen64 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(format.TypeUint64)

	return n + 1, nil
}
