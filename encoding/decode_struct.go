// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"errors"
	"fmt"

	"github.com/basecomplextech/spec/internal/format"
)

func DecodeStruct(b []byte) (dataSize int, size int, err error) {
	if len(b) == 0 {
		return 0, 0, nil
	}

	// Decode type
	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode struct: invalid type")
		return
	}
	if typ != format.TypeStruct {
		err = fmt.Errorf("decode struct: invalid type, type=%v:%d", typ, typ)
		return
	}

	size = n
	end := len(b) - size

	// Data size
	dataSize_, n := decodeSize(b[:end])
	if n < 0 {
		err = errors.New("decode struct: invalid data size")
		return
	}
	size += n + int(dataSize_)

	return int(dataSize_), size, nil
}
