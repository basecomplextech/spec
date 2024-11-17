// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/baselibrary/encoding/compactint"
	"github.com/basecomplextech/spec/internal/core"
)

func EncodeBool(b buffer.Buffer, v bool) (int, error) {
	p := b.Grow(1)
	if v {
		p[0] = byte(core.TypeTrue)
	} else {
		p[0] = byte(core.TypeFalse)
	}
	return 1, nil
}

func EncodeByte(b buffer.Buffer, v byte) (int, error) {
	p := b.Grow(2)
	p[0] = v
	p[1] = byte(core.TypeByte)
	return 2, nil
}

// private

// appendSize appends size as compactint, for tests.
func appendSize(b []byte, big bool, size uint32) []byte {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], size)
	off := compactint.MaxLen32 - n

	return append(b, p[off:]...)
}

func encodeSize(b buffer.Buffer, size uint32) int {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], size)
	off := compactint.MaxLen32 - n

	buf := b.Grow(n)
	copy(buf, p[off:])

	return n
}

func encodeSizeType(b buffer.Buffer, size uint32, type_ core.Type) int {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], size)
	off := compactint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(type_)

	return n + 1
}
