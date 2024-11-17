// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encode

import (
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/format"
)

func EncodeBool(b buffer.Buffer, v bool) (int, error) {
	p := b.Grow(1)
	if v {
		p[0] = byte(format.TypeTrue)
	} else {
		p[0] = byte(format.TypeFalse)
	}
	return 1, nil
}

func EncodeByte(b buffer.Buffer, v byte) (int, error) {
	p := b.Grow(2)
	p[0] = v
	p[1] = byte(format.TypeByte)
	return 2, nil
}
