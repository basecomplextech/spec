// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"encoding/binary"
	"fmt"

	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/core"
)

func EncodeListTable(b buffer.Buffer, dataSize int, table []ListElement) (int, error) {
	if dataSize > core.MaxSize {
		return 0, fmt.Errorf("encode: list too large, max size=%d, actual size=%d", core.MaxSize, dataSize)
	}

	// core.Type
	big := isBigList(table)
	type_ := core.TypeList
	if big {
		type_ = core.TypeBigList
	}

	// Write table
	tableSize, err := encodeListTable(b, table, big)
	if err != nil {
		return int(tableSize), err
	}
	n := tableSize

	// Write data size
	n += encodeSize(b, uint32(dataSize))

	// Write table size and type
	n += encodeSizeType(b, uint32(tableSize), type_)
	return n, nil
}

func encodeListTable(b buffer.Buffer, table []ListElement, big bool) (int, error) {
	// Element size
	elemSize := listElementSmallSize
	if big {
		elemSize = listElementBigSize
	}

	// Check table size
	size := len(table) * elemSize
	if size > core.MaxSize {
		return 0, fmt.Errorf("encode: list table too large, max size=%d, actual size=%d", core.MaxSize, size)
	}

	// Write table
	p := b.Grow(size)
	off := 0

	// Put elements
	for _, elem := range table {
		q := p[off : off+elemSize]

		if big {
			binary.BigEndian.PutUint32(q, elem.Offset)
		} else {
			binary.BigEndian.PutUint16(q, uint16(elem.Offset))
		}

		off += elemSize
	}

	return size, nil
}
