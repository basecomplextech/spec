package protocol

import "fmt"

// entry

type entryType int

const (
	entryTypeUndefined entryType = iota
	entryTypeData
	entryTypeList
	entryTypeElement
	entryTypeMessage
	entryTypeField
)

type entry struct {
	type_ entryType

	// all
	// start offset in data buffer
	start int

	// data
	// end offset in data buffer
	end int

	// list/message
	// table offset in list/message stack
	tableStart int

	// field
	// message field tag
	tag uint16
}

func dataEntry(start int, end int) entry {
	return entry{
		type_: entryTypeData,
		start: start,
		end:   end,
	}
}

func listEntry(start int, tableStart int) entry {
	return entry{
		type_:      entryTypeList,
		start:      start,
		tableStart: tableStart,
	}
}

func listElementEntry(start int) entry {
	return entry{
		type_: entryTypeElement,
		start: start,
	}
}

func messageEntry(start int, tableStart int) entry {
	return entry{
		type_:      entryTypeMessage,
		start:      start,
		tableStart: tableStart,
	}
}

func messageFieldEntry(start int, tag uint16) entry {
	return entry{
		type_: entryTypeField,
		start: start,
		tag:   tag,
	}
}

// stack

type writeStack []entry

// peek returns the top entry on the stack.
func (s writeStack) peek() entry {
	ln := len(s)
	return s[ln-1]
}

// peekType returns the top entry on the stack and checks its type.
func (s writeStack) peekType(type_ entryType) (entry, error) {
	e := s.peek()
	if e.type_ != type_ {
		return e, fmt.Errorf("unexpected stack entry, actual=%v, expected=%v", e.type_, type_)
	}
	return e, nil
}

// pop removes the top entry from the stack.
func (s *writeStack) pop() entry {
	q := *s
	ln := len(q)
	last := q[ln-1]

	*s = q[:ln-1]
	return last
}

// popType removes the top entry from the stack and checks its type.
func (s *writeStack) popType(type_ entryType) (entry, error) {
	e, err := s.peekType(type_)
	if err != nil {
		return e, err
	}

	e = s.pop()
	return e, nil
}

// push pushes a new entry onto the stack.
func (s *writeStack) push(e entry) {
	q := *s

	// realloc
	if cap(q) == len(q) {
		size := cap(q) * 2
		if size == 0 {
			size = 32
		}

		buf := make([]entry, cap(q), size)
		copy(buf, q)
		*s = buf
		q = *s
	}

	q = append(q, e)
	*s = q
}

func (s *writeStack) pushData(start int, end int) {
	e := dataEntry(start, end)
	s.push(e)
}

func (s *writeStack) pushList(start int, tableStart int) {
	e := listEntry(start, tableStart)
	s.push(e)
}

func (s *writeStack) pushElement(start int) {
	e := listElementEntry(start)
	s.push(e)
}

func (s *writeStack) pushMessage(start int, tableStart int) {
	e := messageEntry(start, tableStart)
	s.push(e)
}

func (s *writeStack) pushField(start int, tag uint16) {
	e := messageFieldEntry(start, tag)
	s.push(e)
}

// util

func (t entryType) String() string {
	switch t {
	case entryTypeUndefined:
		return "undefined"
	case entryTypeData:
		return "data"
	case entryTypeList:
		return "list"
	case entryTypeElement:
		return "element"
	case entryTypeMessage:
		return "message"
	case entryTypeField:
		return "field"
	}
	return ""
}
