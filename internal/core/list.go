// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package core

import (
	"encoding/binary"
	"math"
)

const (
	ListElementSize_Small = 2
	ListElementSize_Big   = 4
)

// ListTable is a serialized array of list element offsets ordered by index.
// The serialization format depends on whether the list is big or small, see IsBigList().
//
//	+----------------+----------------+----------------+
//	|    off0(2/4)   |    off1(2/4)   |    off2(2/4)   |
//	+----------------+----------------+----------------+
type ListTable struct {
	table listTable

	data uint32 // data size
	big  bool   // small/big table format
}

// ListElement specifies a value offset in a list byte array.
//
//	+-------------------+
//	|    offset(2/4)    |
//	+-------------------+
type ListElement struct {
	Offset uint32
}

// IsBigList returns true if table count > uint8 or element offset > uint16.
func IsBigList(elements []ListElement) bool {
	ln := len(elements)
	if ln == 0 {
		return false
	}

	// Len > uint8
	if ln > math.MaxUint8 {
		return true
	}

	// Or offset > uint16
	last := elements[ln-1]
	return last.Offset > math.MaxUint16
}

// ListTable

func NewListTable(table listTable, data uint32, big bool) ListTable {
	return ListTable{
		table: table,
		data:  data,
		big:   big,
	}
}

// Len returns the number of elements in the table.
func (t ListTable) Len() int {
	return t.table.len(t.big)
}

// DataSize returns the size of the list data.
func (t ListTable) DataSize() uint32 {
	return t.data
}

// Elements parses the table and returns a slice of elements.
func (t ListTable) Elements() []ListElement {
	return t.table.elements(t.big)
}

// Offset returns an element start/end by an index or -1/-1.
func (t ListTable) Offset(i int) (int, int) {
	if t.big {
		return t.table.offset_big(i)
	} else {
		return t.table.offset_small(i)
	}
}

// internal

type listTable []byte

// len returns the number of elements in the table.
func (t listTable) len(big bool) int {
	var size int
	if big {
		size = ListElementSize_Big
	} else {
		size = ListElementSize_Small
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
	size := ListElementSize_Big
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
	size := ListElementSize_Small
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
