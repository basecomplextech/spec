package spec

// object

type objectType byte

const (
	objectTypeUndefined objectType = iota
	objectTypeList
	objectTypeElement
	objectTypeMessage
	objectTypeField
	objectTypeStruct
)

type objectEntry struct {
	start      int // start offset in data buffer
	tableStart int // table offset in list/message stack
	type_      objectType
}

func (e objectEntry) tag() uint16 {
	return uint16(e.tableStart)
}

type objectStack struct {
	stack []objectEntry
}

func (s *objectStack) reset() {
	s.stack = s.stack[:0]
}

func (s *objectStack) len() int {
	return len(s.stack)
}

// peek returns the last object.
func (s *objectStack) peek() (objectEntry, bool) {
	ln := len(s.stack)
	if ln == 0 {
		return objectEntry{}, false
	}

	e := s.stack[ln-1]
	return e, true
}

// pop removes the top object from the stack and checks its type.
func (s *objectStack) pop() (objectEntry, bool) {
	ln := len(s.stack)
	if ln == 0 {
		return objectEntry{}, false
	}

	e := s.stack[ln-1]
	s.stack = s.stack[:ln-1]
	return e, true
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

func (s *objectStack) pushElement(start int) {
	e := objectEntry{
		type_: objectTypeElement,
		start: start,
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

func (s *objectStack) pushField(start int, tag uint16) {
	e := objectEntry{
		type_:      objectTypeField,
		start:      start,
		tableStart: int(tag),
	}
	s.stack = append(s.stack, e)
}

func (s *objectStack) pushStruct(start int) {
	e := objectEntry{
		type_: objectTypeStruct,
		start: start,
	}
	s.stack = append(s.stack, e)
}

// listStack acts as a buffer for nested list elements.
//
// Each list externally stores its start offset in the buffer, and provides the offset
// when inserting new elements.
//
//	        list0              sublist1            sublist2
//	+-------------------+-------------------+-------------------+
//	| e0 | e1 | e2 | e3 | e0 | e1 | e2 | e3 | e0 | e1 | e2 | e3 |
//	+-------------------+-------------------+-------------------+
//
type listStack struct {
	stack []listElement
}

func (s *listStack) reset() {
	s.stack = s.stack[:0]
}

// offset returns the next list buffer offset.
func (s *listStack) offset() int {
	return len(s.stack)
}

// push appends a new element to the last list.
func (s *listStack) push(elem listElement) {
	s.stack = append(s.stack, elem)
}

// pop pops a list table starting at offset.
func (s *listStack) pop(offset int) []listElement {
	table := s.stack[offset:]
	s.stack = s.stack[:offset]
	return table
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
