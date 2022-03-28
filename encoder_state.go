package spec

import (
	"sort"
	"sync"

	"github.com/baseblck/library/buffer"
)

// encoderState is a big pooled struct which holds an encoding state.
type encoderState struct {
	buf  buffer.Buffer
	data encodeData // last written data, must be consumed before writing next data

	stack    stack
	elements listBuffer    // buffer for list element tables
	fields   messageBuffer // buffer for message field tables

	// preallocated
	_stack    [14]stackEntry
	_elements [48]listElement
	_fields   [48]messageField
}

func newEncoderState() *encoderState {
	s := &encoderState{}
	s.stack.stack = s._stack[:0]
	s.elements.stack = s._elements[:0]
	s.fields.stack = s._fields[:0]
	return s
}

func (s *encoderState) init(b buffer.Buffer) {
	s.reset()
	s.buf = b
}

func (s *encoderState) reset() {
	s.buf = nil
	s.data = encodeData{}

	s.stack.reset()
	s.elements.reset()
	s.fields.reset()
}

// state pool

var encoderStatePool = &sync.Pool{
	New: func() any {
		return newEncoderState()
	},
}

func getEncoderState() *encoderState {
	return encoderStatePool.Get().(*encoderState)
}

func releaseEncoderState(s *encoderState) {
	s.reset()
	encoderStatePool.Put(s)
}

// stack

type entryType byte

const (
	entryUndefined entryType = iota
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

func (e stackEntry) tag() uint16 {
	return uint16(e.tableStart)
}

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

// pop removes the top object from the stack and checks its type.
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

// listBuffer is an encoding buffer for nested list elements.
//
// Each list externally stores its start offset in the buffer, and provides the offset
// when inserting new elements.
//
//	        list0              sublist1            sublist2
//	+-------------------+-------------------+-------------------+
//	| e0 | e1 | e2 | e3 | e0 | e1 | e2 | e3 | e0 | e1 | e2 | e3 |
//	+-------------------+-------------------+-------------------+
//
type listBuffer struct {
	stack []listElement
}

func (b *listBuffer) reset() {
	b.stack = b.stack[:0]
}

// offset returns the next list buffer offset.
func (b *listBuffer) offset() int {
	return len(b.stack)
}

// len returns the number of elements in the last list.
func (b *listBuffer) len(offset int) int {
	table := b.stack[offset:]
	return len(table)
}

// push appends a new element to the last list.
func (b *listBuffer) push(elem listElement) {
	b.stack = append(b.stack, elem)
}

// pop pops a list table starting at offset.
func (b *listBuffer) pop(offset int) []listElement {
	table := b.stack[offset:]
	b.stack = b.stack[:offset]
	return table
}

// messageBuffer is an encoding buffer for nested message fields.
//
// Each message externally stores its start offset in the buffer, and provides the offset
// when inserting new fields. Message fields are kept sorted by tags using the insertion sort.
//
//	       message0          submessage1         submessage2
//	+-------------------+-------------------+-------------------+
//	| f0 | f1 | f2 | f3 | f0 | f1 | f2 | f3 | f0 | f1 | f2 | f3 |
//	+-------------------+-------------------+-------------------+
//
type messageBuffer struct {
	stack []messageField
}

func (b *messageBuffer) reset() {
	b.stack = b.stack[:0]
}

// offset returns the next message table buffer offset.
func (b *messageBuffer) offset() int {
	return len(b.stack)
}

// insert inserts a new field into the last table starting at offset, keeps the table sorted.
func (b *messageBuffer) insert(offset int, f messageField) {
	// append new field
	b.stack = append(b.stack, f)

	// get table
	table := b.stack[offset:]

	// walk table in reverse order
	// move new field to its position
	// using insertion sort
	for i := len(table) - 1; i > 0; i-- {
		left := table[i-1]
		right := table[i]

		// TODO: Replace previous field with the same tag?
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
func (b *messageBuffer) pop(offset int) []messageField {
	table := b.stack[offset:]
	b.stack = b.stack[:offset]
	return table
}

func (b *messageBuffer) hasField(offset int, tag uint16) bool {
	table := b.stack[offset:]

	n := sort.Search(len(table), func(i int) bool {
		return table[i].tag == tag
	})
	if n >= len(table) {
		return false
	}

	return table[n].tag == tag
}
