// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"errors"
	"fmt"

	"github.com/basecomplextech/spec/internal/format"
)

func DecodeMessageTable(b []byte) (_ format.MessageTable, size int, err error) {
	if len(b) == 0 {
		return
	}

	// Decode type
	typ, n := decodeType(b)
	if n < 0 {
		err = errors.New("decode message: invalid type")
		return
	}
	switch typ {
	case format.TypeMessage, format.TypeBigMessage:
	default:
		err = fmt.Errorf("decode message: invalid type, type=%v:%d", typ, typ)
		return
	}

	// Start
	size = n
	end := len(b) - size
	big := typ == format.TypeBigMessage

	// Table size
	tableSize, m := decodeSize(b[:end])
	if m < 0 {
		err = errors.New("decode message: invalid table size")
		return
	}
	end -= m
	size += m

	// Data size
	dataSize, m := decodeSize(b[:end])
	if m < 0 {
		err = fmt.Errorf("decode message: invalid data size")
		return
	}
	end -= m
	size += m

	// Table
	table, err := decodeMessageTable(b[:end], tableSize, big)
	if err != nil {
		return
	}
	end -= int(tableSize) + int(dataSize)
	size += int(tableSize)

	// Data
	if end < 0 {
		err = errors.New("decode message: invalid data")
		return
	}
	size += int(dataSize)

	// Done
	t := format.NewMessageTable(table, dataSize, big)
	return t, size, nil
}

// private

func decodeMessageTable(b []byte, size uint32, big bool) (_ []byte, err error) {
	// Field size
	fieldSize := format.MessageFieldSize_Small
	if big {
		fieldSize = format.MessageFieldSize_Big
	}

	// Check offset
	start := len(b) - int(size)
	if start < 0 {
		err = errors.New("decode message: invalid table")
		return
	}

	// Check divisible
	if size%uint32(fieldSize) != 0 {
		err = errors.New("decode message: invalid table")
		return
	}

	table := b[start:]
	return table, nil
}
