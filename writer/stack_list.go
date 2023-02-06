package writer

import "github.com/complex1tech/spec/encoding"

// listStack is a stack for encoding nested list tables.
//
// Each list externally stores its table offset in the stack,
// and provides it when inserting new elements.
//
//	        list0              sublist1            sublist2
//	+-------------------+-------------------+-------------------+
//	| e0 | e1 | e2 | e3 | e0 | e1 | e2 | e3 | e0 | e1 | e2 | e3 |
//	+-------------------+-------------------+-------------------+
type listStack struct {
	stack []encoding.ListElement
}

func (s *listStack) reset() {
	s.stack = s.stack[:0]
}

// offset returns the next list stack offset.
func (s *listStack) offset() int {
	return len(s.stack)
}

// len returns the number of elements in the last list.
func (s *listStack) len(tableOffset int) int {
	table := s.stack[tableOffset:]
	return len(table)
}

// push appends a new element to the last list.
func (s *listStack) push(elem encoding.ListElement) {
	s.stack = append(s.stack, elem)
}

// pop pops a list table starting at offset.
func (s *listStack) pop(tableOffset int) []encoding.ListElement {
	table := s.stack[tableOffset:]
	s.stack = s.stack[:tableOffset]
	return table
}
