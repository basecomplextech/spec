// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"fmt"

	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/core"
)

func EncodeStruct(b buffer.Buffer, dataSize int) (int, error) {
	if dataSize > core.MaxSize {
		return 0, fmt.Errorf("encode: struct too large, max size=%d, actual size=%d", core.MaxSize, dataSize)
	}

	n := encodeSizeType(b, uint32(dataSize), core.TypeStruct)
	return n, nil
}
