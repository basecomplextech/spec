package protocol

import (
	"encoding/binary"
	"fmt"
)

// elementTable is a serialized array of list element offsets ordered by index.
//
//  +--------------+--------------+--------------+
// 	|    off0(4)   |    off1(4)   |    off2(4)   |
//  +--------------+--------------+--------------+
//
type elementTable []byte

// readElementTable casts bytes into an element table,
// returns an error if length is not divisible by elementSize.
func readElementTable(data []byte) (elementTable, error) {
	ln := len(data)
	if (ln % elementSize) != 0 {
		return nil, fmt.Errorf(
			"read element table: invalid table length, must be divisible by %d, length=%v",
			elementSize, ln,
		)
	}

	return data, nil
}

// writeElementTable writes elements to a binary element table.
// used in tests.
func writeElementTable(elements []element) elementTable {
	// alloc table
	size := len(elements) * elementSize
	result := make([]byte, size)

	// write elements
	for i, elem := range elements {
		off := i * elementSize
		b := result[off:]

		binary.BigEndian.PutUint32(b, uint32(elem.offset))
	}

	return result
}

// get returns an element by its index, panics if index is out of range.
func (t elementTable) get(i int) element {
	n := t.count()
	if i >= n {
		panic(fmt.Sprintf("get element: index out of range, length=%d, index=%d", n, i))
	}

	off := i * elementSize
	b := t[off : off+elementSize]
	elem := element{offset: binary.BigEndian.Uint32(b)}
	return elem
}

// lookup returns an element by its index or false.
func (t elementTable) lookup(i int) (element, bool) {
	n := t.count()
	if i >= n {
		return element{}, false
	}

	off := i * elementSize
	b := t[off : off+elementSize]
	elem := element{offset: binary.BigEndian.Uint32(b)}
	return elem, true
}

// count returns the number of elements in the table.
func (t elementTable) count() int {
	return len(t) / elementSize
}

// elements parses the table and returns a slice of elements
func (t elementTable) elements() []element {
	n := t.count()

	result := make([]element, 0, n)
	for i := 0; i < n; i++ {
		elem := t.get(i)
		result = append(result, elem)
	}
	return result
}
