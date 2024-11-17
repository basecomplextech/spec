// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"encoding/binary"
	"fmt"

	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/format"
)

func EncodeListTable(b buffer.Buffer, dataSize int, table []format.ListElement) (int, error) {
	if dataSize > format.MaxSize {
		return 0, fmt.Errorf("encode: list too large, max size=%d, actual size=%d", format.MaxSize, dataSize)
	}

	// format.Type
	big := format.IsBigList(table)
	type_ := format.TypeList
	if big {
		type_ = format.TypeBigList
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

// private

func encodeListTable(b buffer.Buffer, table []format.ListElement, big bool) (int, error) {
	// Element size
	elemSize := format.ListElementSize_Small
	if big {
		elemSize = format.ListElementSize_Big
	}

	// Check table size
	size := len(table) * elemSize
	if size > format.MaxSize {
		return 0, fmt.Errorf("encode: list table too large, max size=%d, actual size=%d", format.MaxSize, size)
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
