package spec

import (
	"encoding/binary"
	"math"
)

const (
	listElementSmallSize = 2
	listElementBigSize   = 4
)

type listMeta struct {
	table listTable

	data uint32 // data size
	big  bool   // small/big table format
}

// len returns the number of elements in the table.
func (m listMeta) len() int {
	return m.table.len(m.big)
}

// offset returns an element start/end by an index or -1/-1.
func (m listMeta) offset(i int) (int, int) {
	if m.big {
		return m.table.offset_big(i)
	} else {
		return m.table.offset_small(i)
	}
}

// listTable is a serialized array of list element offsets ordered by index.
// the serialization format depends on whether the list is big or small, see isBigList().
//
//	+----------------+----------------+----------------+
//	|    off0(2/4)   |    off1(2/4)   |    off2(2/4)   |
//	+----------------+----------------+----------------+
type listTable []byte

// listElement specifies a value offset in a list byte array.
//
//	+-------------------+
//	|    offset(2/4)    |
//	+-------------------+
type listElement struct {
	offset uint32
}

// isBigList returns true if table count > uint8 or element offset > uint16.
func isBigList(table []listElement) bool {
	ln := len(table)
	if ln == 0 {
		return false
	}

	// len > uint8
	if ln > math.MaxUint8 {
		return true
	}

	// or offset > uint16
	last := table[ln-1]
	return last.offset > math.MaxUint16
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
func (t listTable) elements(big bool) []listElement {
	n := t.len(big)

	result := make([]listElement, 0, n)
	for i := 0; i < n; i++ {
		var end int
		if big {
			_, end = t.offset_big(i)
		} else {
			_, end = t.offset_small(i)
		}

		elem := listElement{
			offset: uint32(end),
		}
		result = append(result, elem)
	}
	return result
}

func (t listTable) offset_big(i int) (int, int) {
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

func (t listTable) offset_small(i int) (int, int) {
	size := listElementSmallSize
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
