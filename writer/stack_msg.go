package writer

import (
	"sort"

	"github.com/basecomplextech/spec/encoding"
)

// messageStack is a stack for encoding nested message tables.
//
// Each message externally stores its table offset in the stack,
// and provides it when inserting new fields.
// Fields are kept sorted by tags using the insertion sort.
//
//	       message0          submessage1         submessage2
//	+-------------------+-------------------+-------------------+
//	| f0 | f1 | f2 | f3 | f0 | f1 | f2 | f3 | f0 | f1 | f2 | f3 |
//	+-------------------+-------------------+-------------------+
type messageStack struct {
	stack []encoding.MessageField
}

func (s *messageStack) reset() {
	s.stack = s.stack[:0]
}

// offset returns the next message table stack offset.
func (s *messageStack) offset() int {
	return len(s.stack)
}

// insert inserts a new field into the last table starting at offset, keeps the table sorted.
func (s *messageStack) insert(tableOffset int, f encoding.MessageField) {
	// Append new field
	s.stack = append(s.stack, f)

	// Get table
	table := s.stack[tableOffset:]

	// Walk table in reverse order
	// Move new field to its position
	// Using insertion sort
	for i := len(table) - 1; i > 0; i-- {
		left := table[i-1]
		right := table[i]

		// TODO: Replace previous field with the same tag?
		if left.Tag < right.Tag {
			// Sorted
			break
		}

		// Swap fields
		table[i-1] = right
		table[i] = left
	}
}

// pop pops a message table starting at offset.
func (s *messageStack) pop(tableOffset int) []encoding.MessageField {
	table := s.stack[tableOffset:]
	s.stack = s.stack[:tableOffset]
	return table
}

func (s *messageStack) hasField(tableOffset int, tag uint16) bool {
	table := s.stack[tableOffset:]

	n := sort.Search(len(table), func(i int) bool {
		return table[i].Tag == tag
	})
	if n >= len(table) {
		return false
	}

	return table[n].Tag == tag
}
