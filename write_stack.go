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

type (
	entry struct {
		type_ entryType

		data    entryData
		list    entryList
		element entryElement
		message entryMessage
		field   entryField
	}

	entryData struct {
		start int
		end   int
	}

	entryList struct {
		start      int
		tableStart int
	}

	entryElement struct {
		start int
	}

	entryMessage struct {
		start      int
		tableStart int
	}

	entryField struct {
		tag   uint16
		start int
	}
)

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
	e := entry{
		type_: entryTypeData,
		data: entryData{
			start: start,
			end:   end,
		},
	}

	s.push(e)
}

func (s *writeStack) pushList(start int, tableStart int) {
	e := entry{
		type_: entryTypeList,

		list: entryList{
			start:      start,
			tableStart: tableStart,
		},
	}
	s.push(e)
}

func (s *writeStack) pushElement(start int) {
	e := entry{
		type_: entryTypeElement,
		element: entryElement{
			start: start,
		},
	}

	s.push(e)
}

func (s *writeStack) pushMessage(start int, tableStart int) {
	e := entry{
		type_: entryTypeMessage,

		message: entryMessage{
			start:      start,
			tableStart: tableStart,
		},
	}

	s.push(e)
}

func (s *writeStack) pushField(tag uint16, start int) {
	e := entry{
		type_: entryTypeField,
		field: entryField{
			tag:   tag,
			start: start,
		},
	}

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
