package spec

import (
	"encoding/binary"
	"math"
)

const (
	messageFieldSize    = 1 + 2 // tag(1) + offset(2)
	messageFieldBigSize = 2 + 4 // tag(2) + offset(4)
)

// isBigList returns true if table count > uint8 or field offset > uint16.
func isBigMessage(table []messageField) bool {
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

// messageField specifies a field tag and a field value offset in message data array.
//
//  +----------+-------------------+
// 	| tag(1/2) |    offset(2/4)    |
//  +----------+-------------------+
//
type messageField struct {
	tag    uint16
	offset uint32
}

// messageTable is a serialized array of message fields ordered by tags.
//
//          field0                field1                field2
//  +---------------------+---------------------+---------------------+
// 	|  tag0 |   offset0   |  tag1 |   offset1   |  tag2 |   offset3   |
//  +---------------------+---------------------+---------------------+
//
type messageTable []byte

// count returns the number of fields in the table.
func (t messageTable) count(big bool) int {
	var size int
	if big {
		size = messageFieldBigSize
	} else {
		size = messageFieldSize
	}
	return len(t) / size
}

// field returns a field by its index or false,
func (t messageTable) field(big bool, i int) (f messageField, ok bool) {
	// inline size
	var size int
	if big {
		size = messageFieldBigSize
	} else {
		size = messageFieldSize
	}

	// count
	n := len(t) / size
	switch {
	case i < 0:
		return
	case i >= n:
		return
	}

	off := i * size
	b := t[off : off+size]

	if big {
		f = messageField{
			tag:    binary.BigEndian.Uint16(b),
			offset: binary.BigEndian.Uint32(b[2:]),
		}
	} else {
		f = messageField{
			tag:    uint16(b[0]),
			offset: uint32(binary.BigEndian.Uint16(b[1:])),
		}
	}

	ok = true
	return
}

// fields parses the table and returns a slice of fields.
func (t messageTable) fields(big bool) []messageField {
	n := t.count(big)

	result := make([]messageField, 0, n)
	for i := 0; i < n; i++ {
		field, ok := t.field(big, i)
		if !ok {
			continue
		}
		result = append(result, field)
	}
	return result
}

// offset returns field start/end by its tag or -1/-1.
func (t messageTable) offset(big bool, tag uint16) (int, int) {
	// inline size
	var size int
	if big {
		size = messageFieldBigSize
	} else {
		size = messageFieldSize
	}

	// count
	n := len(t) / size

	// binary search table
	left, right := 0, (n - 1)
	for left <= right {
		// middle
		middle := int(uint(left+right) >> 1) // avoid overflow

		// offset
		off := middle * size
		b := t[off : off+size]

		// current tag
		var cur uint16
		if big {
			cur = binary.BigEndian.Uint16(b)
		} else {
			cur = uint16(b[0])
		}

		// check current
		switch {
		case cur < tag:
			left = middle + 1

		case cur > tag:
			right = middle - 1

		case cur == tag:
			var start int
			var end int

			// start
			if middle > 0 {
				if big {
					start = int(binary.BigEndian.Uint32(t[off-4:]))
				} else {
					start = int(binary.BigEndian.Uint16(t[off-2:]))
				}
			}

			// end
			if big {
				end = int(binary.BigEndian.Uint32(b[2:]))
			} else {
				end = int(binary.BigEndian.Uint16(b[1:]))
			}
			return start, end
		}
	}

	return -1, -1
}

// offsetByIndex returns field start/end by its index or -1/-1.
func (t messageTable) offsetByIndex(big bool, i int) (int, int) {
	// inline size
	var size int
	if big {
		size = messageFieldBigSize
	} else {
		size = messageFieldSize
	}

	// count
	n := len(t) / size
	switch {
	case i < 0:
		return -1, -1
	case i >= n:
		return -1, -1
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
		end = int(binary.BigEndian.Uint32(t[off+2:]))
	} else {
		end = int(binary.BigEndian.Uint16(t[off+1:]))
	}

	return start, end
}

// messageStack acts as a buffer for nested message fields.
//
// Each message externally stores its start offset in the buffer, and provides the offset
// when inserting new fields. Message fields are kept sorted by tags using the insertion sort.
//
//	       message0          submessage1         submessage2
//	+-------------------+-------------------+-------------------+
//	| f0 | f1 | f2 | f3 | f0 | f1 | f2 | f3 | f0 | f1 | f2 | f3 |
//	+-------------------+-------------------+-------------------+
//
type messageStack struct {
	stack []messageField
}

func (s *messageStack) reset() {
	s.stack = s.stack[:0]
}

// offset returns the next message table buffer offset.
func (s *messageStack) offset() int {
	return len(s.stack)
}

// insert inserts a new field into the last table starting at offset, keeps the table sorted.
func (s *messageStack) insert(offset int, f messageField) {
	// append new field
	s.stack = append(s.stack, f)

	// get table
	table := s.stack[offset:]

	// walk table in reverse order
	// move new field to its position
	// using insertion sort
	for i := len(table) - 1; i > 0; i-- {
		left := table[i-1]
		right := table[i]

		if left.tag < right.tag {
			// sorted
			break
		}

		// swap fields
		table[i-1] = right
		table[i] = left
	}
}

// pop pops a message table starting at offset.
func (s *messageStack) pop(offset int) []messageField {
	table := s.stack[offset:]
	s.stack = s.stack[:offset]
	return table
}
