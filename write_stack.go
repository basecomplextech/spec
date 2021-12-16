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
}

func dataEntry(start int, end int) entry {
	return entry{
		type_: entryTypeData,
		start: start,
		// end:   end,
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

// stack

type writeStack struct {
	stack []entry
}

func (s *writeStack) reset() {
	s.stack = s.stack[:0]
}

func (s *writeStack) len() int {
	return len(s.stack)
}

// peek returns the top entry on the stack and checks its type.
func (s *writeStack) peek(type_ entryType) (entry, error) {
	ln := len(s.stack)
	if ln == 0 {
		return entry{}, fmt.Errorf("peek: stack is empty")
	}

	e := s.stack[ln-1]
	if e.type_ != type_ {
		return e, fmt.Errorf("peek: unexpected stack entry, expected=%v, actual=%v, ", type_, e.type_)
	}
	return e, nil
}

// pop removes the top entry from the stack and checks its type.
func (s *writeStack) pop(type_ entryType) (entry, error) {
	ln := len(s.stack)
	if ln == 0 {
		return entry{}, fmt.Errorf("pop: stack is empty")
	}

	e := s.stack[ln-1]
	if e.type_ != type_ {
		return e, fmt.Errorf("peek: unexpected stack entry, expected=%v, actual=%v, ", type_, e.type_)
	}

	s.stack = s.stack[:ln-1]
	return e, nil
}

// push pushes a new entry onto the stack.
func (s *writeStack) push(e entry) {
	s.stack = append(s.stack, e)
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

// func (s *writeStack) pushField(start int, tag uint16) {
// 	e := messageFieldEntry(start, tag)
// 	s.push(e)
// }

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
