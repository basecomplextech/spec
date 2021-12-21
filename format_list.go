package spec

import (
	"encoding/binary"
	"math"
)

const (
	listElementSize    = 2
	listElementBigSize = 4
)

// isBigList returns true if a table exceeds a standard list count or offset.
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

// offset returns an element start/end by its index or -1/-1.
func (t listTable) offset(big bool, i int) (int, int) {
	// inline count
	var n int
	if big {
		n = len(t) / listElementBigSize
	} else {
		n = len(t) / listElementSize
	}

	// check count
	switch {
	case i < 0:
		return -1, -1
	case i >= n:
		return -1, -1
	}

	// size
	var size int
	if big {
		size = listElementBigSize
	} else {
		size = listElementSize
	}

	// offsets
	off := i * size
	var start int
	var end int

	// start
	if i > 0 {
		if big {
			start = int(binary.BigEndian.Uint32(t[off-4:]))
		} else {
			start = int(binary.BigEndian.Uint16(t[off-2:]))
		}
	}

	// end
	if big {
		end = int(binary.BigEndian.Uint32(t[off:]))
	} else {
		end = int(binary.BigEndian.Uint16(t[off:]))
	}

	return start, end
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

// listStack acts as a buffer for nested list elements.
//
// Each list externally stores its start offset in the buffer, and provides the offset
// when inserting new elements.
//
//	        list0              sublist1            sublist2
//	+-------------------+-------------------+-------------------+
//	| e0 | e1 | e2 | e3 | e0 | e1 | e2 | e3 | e0 | e1 | e2 | e3 |
//	+-------------------+-------------------+-------------------+
//
type listStack struct {
	stack []listElement
}

func (s *listStack) reset() {
	s.stack = s.stack[:0]
}

// offset returns the next list buffer offset.
func (s *listStack) offset() int {
	return len(s.stack)
}

// push appends a new element to the last list.
func (s *listStack) push(elem listElement) {
	s.stack = append(s.stack, elem)
}

// pop pops a list table starting at offset.
func (s *listStack) pop(offset int) []listElement {
	table := s.stack[offset:]
	s.stack = s.stack[:offset]
	return table
}
