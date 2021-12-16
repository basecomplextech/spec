package protocol

import (
	"encoding/binary"
	"fmt"
	"sort"
)

type Message struct {
	bytes []byte

	type_     Type
	tableSize uint32
	dataSize  uint32
	table     messageTable
	data      messageData
}

func ReadMessage(p []byte) Message {
	buf := buffer(p)

	type_, b := buf.type_()
	if type_ != TypeMessage {
		return Message{}
	}

	tableSize, b := b.messageTableSize()
	dataSize, b := b.messageDataSize()
	table, b := b.messageTable(tableSize)
	data, _ := b.messageData(dataSize)
	bytes, _ := buf.messageBytes(tableSize, dataSize) // slice initial buffer

	return Message{
		bytes: bytes,

		type_:     type_,
		tableSize: tableSize,
		dataSize:  dataSize,
		table:     table,
		data:      data,
	}
}

// Bytes returns message bytes.
func (m Message) Bytes() []byte {
	return m.bytes
}

// Field returns a field value by a tag or an empty value.
func (m Message) Field(tag uint16) (Value, bool) {
	field, ok := m.table.lookup(tag)
	if !ok {
		return Value{}, false
	}

	r := m.data.field(field.offset)
	return r, true
}

// FieldByIndex returns a field value by an index or an empty value.
func (m Message) FieldByIndex(i int) (Value, bool) {
	return Value{}, false
}

// Fields returns the number of fields in the message.
func (m Message) Fields() int {
	return m.table.count()
}

// internal

// messageField specifies a field tag and a field value offset in message data array.
//
//  +--------+-----------------+
// 	| tag(2) |    offset(4)    |
//  +--------+-----------------+
//
type messageField struct {
	tag    uint16
	offset uint32
}

const messageFieldSize = 2 + 4 // tag(2) + offset(4)

// messageTable is a serialized array of message fields ordered by tags.
//
//  +---------------------+---------------------+---------------------+
//  |       field0        |       field1        |       field2        |
//  +---------------------+---------------------+---------------------+
// 	|  tag0 |   offset0   |  tag1 |   offset1   |  tag2 |   offset3   |
//  +---------------------+---------------------+---------------------+
//
type messageTable []byte

// readMessageTable casts bytes into a field table,
// returns an error if length is not divisible by messageFieldSize.
func readMessageTable(data []byte) (messageTable, error) {
	ln := len(data)
	if (ln % messageFieldSize) != 0 {
		return nil, fmt.Errorf(
			"read field table: invalid table length, must be divisible by %d, length=%v",
			messageFieldSize, ln,
		)
	}

	return data, nil
}

// writeMessageTable sorts the fields and writes them to a binary field table.
// used in tests.
func writeMessageTable(fields []messageField) messageTable {
	// sort fields
	sort.Slice(fields, func(i, j int) bool {
		a, b := fields[i], fields[j]
		return a.tag < b.tag
	})

	// alloc table
	size := len(fields) * messageFieldSize
	result := make([]byte, size)

	// write sorted fields
	for i, field := range fields {
		off := i * messageFieldSize
		b := result[off:]

		binary.BigEndian.PutUint16(b, field.tag)
		binary.BigEndian.PutUint32(b[2:], field.offset)
	}

	return result
}

// get returns a field by its index or false.
func (t messageTable) get(i int) (messageField, bool) {
	n := t.count()
	switch {
	case i < 0:
		return messageField{}, false
	case i >= n:
		return messageField{}, false
	}

	off := i * messageFieldSize
	b := t[off : off+messageFieldSize]

	f := messageField{
		tag:    binary.BigEndian.Uint16(b),
		offset: binary.BigEndian.Uint32(b[2:]),
	}
	return f, true
}

// find finds a field by a tag and returns its index or -1.
func (t messageTable) find(tag uint16) int {
	n := t.count()
	if n == 0 {
		return -1
	}

	// binary search table
	left, right := 0, (n - 1)
	for left <= right {
		// calc offset
		middle := (left + right) / 2
		off := middle * messageFieldSize

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
func (t messageTable) lookup(tag uint16) (messageField, bool) {
	i := t.find(tag)
	if i < 0 {
		return messageField{}, false
	}

	return t.get(i)
}

// count returns the number of fields in the table.
func (t messageTable) count() int {
	return len(t) / messageFieldSize
}

// fields parses the table and returns a slice of fields.
func (t messageTable) fields() []messageField {
	n := t.count()

	result := make([]messageField, 0, n)
	for i := 0; i < n; i++ {
		field, ok := t.get(i)
		if !ok {
			continue
		}
		result = append(result, field)
	}
	return result
}

// messageData holds message field values referenced by offsets.
//
//  +----------+----------+----------+----------+
// 	|  value0  |  value1  |  value2  |  value3  |
//  +----------+----------+----------+----------+
//
type messageData struct {
	data []byte
}

// field returns a field value by offset or an empty value.
func (d messageData) field(off uint32) Value {
	if len(d.data) < int(off) {
		return Value{}
	}
	b := d.data[:off]
	return ReadValue(b)
}

// messageStack acts as a buffer for nested message fields.
//
// Each message externally stores its start offset in the buffer, and provides the offset
// when inserting new fields. Message fields are kept sorted by tags using the insertion sort.
//
// 	+-------------------+-------------------+-------------------+
//	|      message0     |    submessage1    |    submessage2    |
//	|-------------------|-------------------|-------------------|
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
