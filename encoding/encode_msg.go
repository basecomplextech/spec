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

func EncodeMessageTable(b buffer.Buffer, dataSize int, table []format.MessageField) (int, error) {
	if dataSize > format.MaxSize {
		return 0, fmt.Errorf("encode: message too large, max size=%d, actual size=%d", format.MaxSize, dataSize)
	}

	// format.Type
	big := format.IsBigMessage(table)
	type_ := format.TypeMessage
	if big {
		type_ = format.TypeBigMessage
	}

	// Write table
	tableSize, err := encodeMessageTable(b, table, big)
	if err != nil {
		return 0, err
	}
	n := tableSize

	// Write data size
	n += encodeSize(b, uint32(dataSize))

	// Write table size and type
	n += encodeSizeType(b, uint32(tableSize), type_)
	return n, nil
}

func encodeMessageTable(b buffer.Buffer, table []format.MessageField, big bool) (int, error) {
	// Field size
	var fieldSize int
	if big {
		fieldSize = format.MessageFieldSize_Big
	} else {
		fieldSize = format.MessageFieldSize_Small
	}

	// Check table size
	size := len(table) * fieldSize
	if size > format.MaxSize {
		return 0, fmt.Errorf("encode: message table too large, max size=%d, actual size=%d", format.MaxSize, size)
	}

	// Write table
	p := b.Grow(size)
	off := 0

	// Put fields
	for _, field := range table {
		q := p[off : off+fieldSize]

		if big {
			binary.BigEndian.PutUint16(q, field.Tag)
			binary.BigEndian.PutUint32(q[2:], field.Offset)
		} else {
			q[0] = byte(field.Tag)
			binary.BigEndian.PutUint16(q[1:], uint16(field.Offset))
		}

		off += fieldSize
	}

	return size, nil
}
