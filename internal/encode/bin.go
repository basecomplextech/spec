// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encode

import (
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/format"
)

func EncodeBin64(b buffer.Buffer, v bin.Bin64) (int, error) {
	p := b.Grow(9)
	v.MarshalTo(p)
	p[8] = byte(format.TypeBin64)
	return 9, nil
}

func EncodeBin128(b buffer.Buffer, v bin.Bin128) (int, error) {
	p := b.Grow(17)
	v.MarshalTo(p)
	p[16] = byte(format.TypeBin128)
	return 17, nil
}

func EncodeBin128Bytes(b buffer.Buffer, v bin.Bin128) ([]byte, int, error) {
	p := b.Grow(17)
	v.MarshalTo(p)
	p[16] = byte(format.TypeBin128)
	return p, 17, nil
}

func EncodeBin256(b buffer.Buffer, v bin.Bin256) (int, error) {
	p := b.Grow(33)
	v.MarshalTo(p)
	p[32] = byte(format.TypeBin256)
	return 33, nil
}
