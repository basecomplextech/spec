// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"errors"
	"fmt"

	"github.com/basecomplextech/spec/internal/format"
)

func DecodeListTable(b []byte) (_ format.ListTable, size int, err error) {
	if len(b) == 0 {
		return
	}

	// Decode type
	typ, n := decodeType(b)
	if n < 0 {
		n = 0
		err = errors.New("decode list: invalid data")
		return
	}
	if typ != format.TypeList && typ != format.TypeBigList {
		err = fmt.Errorf("decode list: invalid type, type=%v:%d", typ, typ)
		return
	}

	// Start
	size = n
	end := len(b) - n
	big := typ == format.TypeBigList

	// Table size
	tableSize, n := decodeSize(b[:end])
	if n < 0 {
		err = errors.New("decode list: invalid table size")
		return
	}
	end -= n
	size += n

	// Data size
	dataSize, n := decodeSize(b[:end])
	if n < 0 {
		err = errors.New("decode list: invalid data size")
		return
	}
	end -= n
	size += n

	// Table
	table, err := decodeListTable(b[:end], tableSize, big)
	if err != nil {
		return
	}
	end -= int(tableSize) + int(dataSize)
	size += int(tableSize)

	// Data
	if end < 0 {
		err = errors.New("decode list: invalid data")
		return
	}
	size += int(dataSize)

	// Done
	t := format.NewListTable(table, dataSize, big)
	return t, size, nil
}

// private

func decodeListTable(b []byte, size uint32, big bool) (_ []byte, err error) {
	// Element size
	elemSize := format.ListElementSize_Small
	if big {
		elemSize = format.ListElementSize_Big
	}

	// Check offset
	start := len(b) - int(size)
	if start < 0 {
		err = errors.New("decode list: invalid table")
		return
	}

	// Check divisible
	if size%uint32(elemSize) != 0 {
		err = errors.New("decode list: invalid table")
		return
	}

	table := b[start:]
	return table, nil
}
