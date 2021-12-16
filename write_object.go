package protocol

import "fmt"

// object

type objectType byte

const (
	objectTypeUndefined objectType = iota
	objectTypeList
	objectTypeMessage
)

type objectEntry struct {
	start      int // start offset in data buffer
	tableStart int // table offset in list/message stack
	type_      objectType
}

// stack

type objectStack struct {
	stack []objectEntry
}

func (s *objectStack) reset() {
	s.stack = s.stack[:0]
}

func (s *objectStack) len() int {
	return len(s.stack)
}

// last

// last returns the last object and checks its type.
func (s *objectStack) last(type_ objectType) (objectEntry, error) {
	ln := len(s.stack)
	if ln == 0 {
		return objectEntry{}, fmt.Errorf("last: object stack is empty")
	}

	e := s.stack[ln-1]
	if e.type_ != type_ {
		return e, fmt.Errorf("last: unexpected stack object, expected=%v, actual=%v, ",
			objectTypeList, e.type_)
	}
	return e, nil
}

func (s *objectStack) lastList() (objectEntry, error) {
	return s.last(objectTypeList)
}

func (s *objectStack) lastMessage() (objectEntry, error) {
	return s.last(objectTypeMessage)
}

// pop

// pop removes the top object from the stack and checks its type.
func (s *objectStack) pop(type_ objectType) (objectEntry, error) {
	ln := len(s.stack)
	if ln == 0 {
		return objectEntry{}, fmt.Errorf("pop: stack is empty")
	}

	e := s.stack[ln-1]
	if e.type_ != type_ {
		return e, fmt.Errorf("peek: unexpected object, expected=%v, actual=%v, ", type_, e.type_)
	}

	s.stack = s.stack[:ln-1]
	return e, nil
}

func (s *objectStack) popList() (objectEntry, error) {
	return s.pop(objectTypeList)
}

func (s *objectStack) popMessage() (objectEntry, error) {
	return s.pop(objectTypeMessage)
}

// push

func (s *objectStack) pushList(start int, tableStart int) {
	e := objectEntry{
		type_:      objectTypeList,
		start:      start,
		tableStart: tableStart,
	}
	s.stack = append(s.stack, e)
}

func (s *objectStack) pushMessage(start int, tableStart int) {
	e := objectEntry{
		type_:      objectTypeMessage,
		start:      start,
		tableStart: tableStart,
	}
	s.stack = append(s.stack, e)
}

// util

func (t objectType) String() string {
	switch t {
	case objectTypeUndefined:
		return "undefined"
	case objectTypeList:
		return "list"
	case objectTypeMessage:
		return "message"
	}
	return ""
}
