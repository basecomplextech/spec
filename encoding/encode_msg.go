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

func EncodeMessageTable(b buffer.Buffer, dataSize int, table []core.MessageField) (int, error) {
	if dataSize > core.MaxSize {
		return 0, fmt.Errorf("encode: message too large, max size=%d, actual size=%d", core.MaxSize, dataSize)
	}

	// core.Type
	big := core.IsBigMessage(table)
	type_ := core.TypeMessage
	if big {
		type_ = core.TypeBigMessage
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

func encodeMessageTable(b buffer.Buffer, table []core.MessageField, big bool) (int, error) {
	// Field size
	var fieldSize int
	if big {
		fieldSize = core.MessageFieldSize_Big
	} else {
		fieldSize = core.MessageFieldSize_Small
	}

	// Check table size
	size := len(table) * fieldSize
	if size > core.MaxSize {
		return 0, fmt.Errorf("encode: message table too large, max size=%d, actual size=%d", core.MaxSize, size)
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
