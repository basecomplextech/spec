// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"fmt"

	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/core"
)

func EncodeBytes(b buffer.Buffer, v []byte) (int, error) {
	size := len(v)
	if size > core.MaxSize {
		return 0, fmt.Errorf("encode: bytes too large, max size=%d, actual size=%d", core.MaxSize, size)
	}

	p := b.Grow(size)
	copy(p, v)
	n := size

	n += encodeSizeType(b, uint32(size), core.TypeBytes)
	return n, nil
}
