// Copyright 2023 Ivan Korobkov. All rights reserved.

package writer

type entryType byte

const (
	entryUndefined entryType = iota
	entryData                // data holds the last written data start/end
	entryList
	entryElement
	entryMessage
	entryField
)

type stackEntry struct {
	start      int // start offset in data buffer
	tableStart int // table offset in list/message stack
	type_      entryType
}

func (e stackEntry) end() int {
	return e.tableStart
}

func (e stackEntry) tag() uint16 {
	return uint16(e.tableStart)
}

// stack

type stack struct {
	stack []stackEntry
}

func (s *stack) reset() {
	s.stack = s.stack[:0]
}

func (s *stack) len() int {
	return len(s.stack)
}

// peek returns the last object.
func (s *stack) peek() (stackEntry, bool) {
	ln := len(s.stack)
	if ln == 0 {
		return stackEntry{}, false
	}

	e := s.stack[ln-1]
	return e, true
}

// peekSecondLast returns the second last object.
func (s *stack) peekSecondLast() (stackEntry, bool) {
	ln := len(s.stack)
	if ln < 2 {
		return stackEntry{}, false
	}

	e := s.stack[ln-2]
	return e, true
}

// pop removes the top object from the stack.
func (s *stack) pop() (stackEntry, bool) {
	ln := len(s.stack)
	if ln == 0 {
		return stackEntry{}, false
	}

	e := s.stack[ln-1]
	s.stack = s.stack[:ln-1]
	return e, true
}

// push

func (s *stack) pushData(start, end int) {
	e := stackEntry{
		type_:      entryData,
		start:      start,
		tableStart: end, // end
	}
	s.stack = append(s.stack, e)
}

func (s *stack) pushList(start int, tableStart int) {
	e := stackEntry{
		type_:      entryList,
		start:      start,
		tableStart: tableStart,
	}
	s.stack = append(s.stack, e)
}

func (s *stack) pushElement(start int) {
	e := stackEntry{
		type_: entryElement,
		start: start,
	}
	s.stack = append(s.stack, e)
}

func (s *stack) pushMessage(start int, tableStart int) {
	e := stackEntry{
		type_:      entryMessage,
		start:      start,
		tableStart: tableStart,
	}
	s.stack = append(s.stack, e)
}

func (s *stack) pushField(start int, tag uint16) {
	e := stackEntry{
		type_:      entryField,
		start:      start,
		tableStart: int(tag),
	}
	s.stack = append(s.stack, e)
}
