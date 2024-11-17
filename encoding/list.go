// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"encoding/binary"
	"math"
)

const (
	listElementSmallSize = 2
	listElementBigSize   = 4
)

type ListTable struct {
	table listTable

	data uint32 // data size
	big  bool   // small/big table format
}

// Len returns the number of elements in the table.
func (t ListTable) Len() int {
	return t.table.len(t.big)
}

// DataSize returns the size of the list data.
func (t ListTable) DataSize() uint32 {
	return t.data
}

// Offset returns an element start/end by an index or -1/-1.
func (t ListTable) Offset(i int) (int, int) {
	if t.big {
		return t.table.offset_big(i)
	} else {
		return t.table.offset_small(i)
	}
}

// ListElement specifies a value offset in a list byte array.
//
//	+-------------------+
//	|    offset(2/4)    |
//	+-------------------+
type ListElement struct {
	Offset uint32
}

// listTable is a serialized array of list element offsets ordered by index.
// the serialization format depends on whether the list is big or small, see isBigList().
//
//	+----------------+----------------+----------------+
//	|    off0(2/4)   |    off1(2/4)   |    off2(2/4)   |
//	+----------------+----------------+----------------+
type listTable []byte

// isBigList returns true if table count > uint8 or element offset > uint16.
func isBigList(table []ListElement) bool {
	ln := len(table)
	if ln == 0 {
		return false
	}

	// Len > uint8
	if ln > math.MaxUint8 {
		return true
	}

	// Or offset > uint16
	last := table[ln-1]
	return last.Offset > math.MaxUint16
}

// len returns the number of elements in the table.
func (t listTable) len(big bool) int {
	var size int
	if big {
		size = listElementBigSize
	} else {
		size = listElementSmallSize
	}
	return len(t) / size
}

// elements parses the table and returns a slice of elements
func (t listTable) elements(big bool) []ListElement {
	n := t.len(big)

	result := make([]ListElement, 0, n)
	for i := 0; i < n; i++ {
		var end int
		if big {
			_, end = t.offset_big(i)
		} else {
			_, end = t.offset_small(i)
		}

		elem := ListElement{
			Offset: uint32(end),
		}
		result = append(result, elem)
	}
	return result
}

func (t listTable) offset_big(i int) (int, int) {
	size := listElementBigSize
	n := len(t) / size

	// Check count
	switch {
	case i < 0:
		return -1, -1
	case i >= n:
		return -1, -1
	}

	// Offset
	off := i * size

	// Start
	var start int
	if i > 0 {
		start = int(binary.BigEndian.Uint32(t[off-4:]))
	}

	// End
	end := int(binary.BigEndian.Uint32(t[off:]))
	return start, end
}

func (t listTable) offset_small(i int) (int, int) {
	size := listElementSmallSize
	n := len(t) / size

	// Check count
	switch {
	case i < 0:
		return -1, -1
	case i >= n:
		return -1, -1
	}

	// Offset
	off := i * size

	// Start
	var start int
	if i > 0 {
		start = int(binary.BigEndian.Uint16(t[off-2:]))
	}

	// End
	end := int(binary.BigEndian.Uint16(t[off:]))
	return start, end
}
