package protocol

import (
	"encoding/binary"
	"fmt"
	"sort"
)

// fieldTable is a serialized array of message fields sorted by tags.
//
//  +--------------------+--------------------+--------------------+
// 	| tag0 |   offset0   | tag1 |   offset1   | tag2 |   offset3   |
//  +--------------------+--------------------+--------------------+
//
type fieldTable []byte

// readFieldTable casts bytes into a field table, returns an error if length is not divisible by fieldSize.
func readFieldTable(data []byte) (fieldTable, error) {
	ln := len(data)
	if (ln % fieldSize) != 0 {
		return nil, fmt.Errorf("read field table: invalid length, must be divisible by %d, length=%v",
			fieldSize, ln)
	}

	return data, nil
}

// writeFieldTable sorts the fields and writes them to a binary field table.
// used in tests.
func writeFieldTable(fields []field) fieldTable {
	// sort fields
	sort.Slice(fields, func(i, j int) bool {
		a, b := fields[i], fields[j]
		return a.tag < b.tag
	})

	// alloc table
	size := len(fields) * fieldSize
	result := make([]byte, size)

	// write sorted fields
	for i, field := range fields {
		off := i * fieldSize
		b := result[off:]

		binary.BigEndian.PutUint16(b, field.tag)
		binary.BigEndian.PutUint32(b[2:], field.offset)
	}

	return result
}

// get returns a field by its index, panics if i is out of range.
func (t fieldTable) get(i int) field {
	n := t.count()
	if i >= n {
		panic("get field: index out of range")
	}

	off := i * fieldSize
	b := t[off : off+fieldSize]

	field := field{
		tag:    binary.BigEndian.Uint16(b),
		offset: binary.BigEndian.Uint32(b[2:]),
	}
	return field
}

// find finds a field by a tag and returns its index or -1.
func (t fieldTable) find(tag uint16) int {
	n := t.count()
	if n == 0 {
		return -1
	}

	left, right := 0, (n - 1)
	for left <= right {
		// calc offset
		middle := (left + right) / 2
		off := middle * fieldSize

		// read current tag
		cur := binary.BigEndian.Uint16(t[off:])

		// check current
		switch {
		case cur == tag:
			return middle
		case cur < tag:
			left = middle + 1
		case cur > tag:
			right = middle - 1
		}
	}
	return -1
}

// lookup finds and returns a field by a tag, or returns false.
func (t fieldTable) lookup(tag uint16) (field, bool) {
	i := t.find(tag)
	if i < 0 {
		return field{}, false
	}

	f := t.get(i)
	return f, true
}

// count returns the number of fields in the table.
func (t fieldTable) count() int {
	return len(t) / fieldSize
}

// fields parses the table and returns a slice of fields.
func (t fieldTable) fields() []field {
	n := t.count()

	result := make([]field, 0, n)
	for i := 0; i < n; i++ {
		field := t.get(i)
		result = append(result, field)
	}
	return result
}
