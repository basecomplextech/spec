package spec

import (
	"encoding/binary"
	"math"
)

const (
	listElementSize    = 2
	listElementBigSize = 4
)

// isBigList returns true if table count > uint8 or element offset > uint16.
func isBigList(table []listElement) bool {
	ln := len(table)
	if ln == 0 {
		return false
	}

	// count > uint8
	if ln > math.MaxUint8 {
		return true
	}

	// or offset > uint16
	last := table[ln-1]
	return last.offset > math.MaxUint16
}

// listElement specifies an element value offset in list data array.
//
//	+-------------------+
//	|    offset(2/4)    |
//	+-------------------+
//
type listElement struct {
	offset uint32
}

// listTable is a serialized array of list element offsets ordered by index.
//
//	+----------------+----------------+----------------+
//	|    off0(2/4)   |    off1(2/4)   |    off2(2/4)   |
//	+----------------+----------------+----------------+
//
type listTable []byte

// count returns the number of elements in the table.
func (t listTable) count(big bool) int {
	var size int
	if big {
		size = listElementBigSize
	} else {
		size = listElementSize
	}
	return len(t) / size
}

// elements parses the table and returns a slice of elements
func (t listTable) elements(big bool) []listElement {
	n := t.count(big)

	result := make([]listElement, 0, n)
	for i := 0; i < n; i++ {
		_, end := t.offset(big, i)

		elem := listElement{
			offset: uint32(end),
		}
		result = append(result, elem)
	}
	return result
}

// offset returns an element start/end by its index or -1/-1.
func (t listTable) offset(big bool, i int) (int, int) {
	if big {
		return t._offset_big(i)
	} else {
		return t._offset_small(i)
	}
}

func (t listTable) _offset_big(i int) (int, int) {
	size := listElementBigSize
	n := len(t) / size

	// check count
	switch {
	case i < 0:
		return -1, -1
	case i >= n:
		return -1, -1
	}

	// offset
	off := i * size

	// start
	var start int
	if i > 0 {
		start = int(binary.BigEndian.Uint32(t[off-4:]))
	}

	// end
	end := int(binary.BigEndian.Uint32(t[off:]))
	return start, end
}

func (t listTable) _offset_small(i int) (int, int) {
	size := listElementSize
	n := len(t) / size

	// check count
	switch {
	case i < 0:
		return -1, -1
	case i >= n:
		return -1, -1
	}

	// offset
	off := i * size

	// start
	var start int
	if i > 0 {
		start = int(binary.BigEndian.Uint16(t[off-2:]))
	}

	// end
	end := int(binary.BigEndian.Uint16(t[off:]))
	return start, end
}
