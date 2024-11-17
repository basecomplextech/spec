// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encode

import (
	"fmt"

	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/format"
)

func EncodeString(b buffer.Buffer, s string) (int, error) {
	size := len(s)
	if size > format.MaxSize {
		return 0, fmt.Errorf("encode: string too large, max size=%d, actual size=%d", format.MaxSize, size)
	}

	n := size + 1 // plus zero byte
	p := b.Grow(n)
	copy(p, s)

	n += encodeSizeType(b, uint32(size), format.TypeString)
	return n, nil
}
