// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encode

import (
	"fmt"

	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/format"
)

func EncodeStruct(b buffer.Buffer, dataSize int) (int, error) {
	if dataSize > format.MaxSize {
		return 0, fmt.Errorf("encode: struct too large, max size=%d, actual size=%d", format.MaxSize, dataSize)
	}

	n := encodeSizeType(b, uint32(dataSize), format.TypeStruct)
	return n, nil
}
