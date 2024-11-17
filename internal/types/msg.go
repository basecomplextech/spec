// Copyright 2022 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package types

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/decode"
	"github.com/basecomplextech/spec/internal/format"
)

// Message is a raw message.
type Message struct {
	table format.MessageTable
	bytes []byte
}

// MessageType is a type implemented by generated messages.
type MessageType interface {
	Unwrap() Message
}

// OpenMessage opens and returns a message from bytes, or an empty message on error.
// The method decodes the message table, but not the fields.
func OpenMessage(b []byte) Message {
	table, n, err := decode.DecodeMessageTable(b)
	if err != nil {
		return Message{}
	}
	bytes := b[len(b)-n:]

	return Message{
		table: table,
		bytes: bytes,
	}
}

// OpenMessageErr opens and returns a message from bytes, or an error.
// The method decodes the message table, but not the fields.
func OpenMessageErr(b []byte) (Message, error) {
	table, size, err := decode.DecodeMessageTable(b)
	if err != nil {
		return Message{}, err
	}
	bytes := b[len(b)-size:]

	m := Message{
		table: table,
		bytes: bytes,
	}
	return m, nil
}

// ParseMessage recursively parses and returns a message.
func ParseMessage(b []byte) (_ Message, size int, err error) {
	table, size, err := decode.DecodeMessageTable(b)
	if err != nil {
		return Message{}, 0, err
	}
	bytes := b[len(b)-size:]

	m := Message{
		table: table,
		bytes: bytes,
	}

	num := m.Fields()
	for i := 0; i < num; i++ {
		b1 := m.fieldAt(i)
		if len(b1) == 0 {
			continue
		}

		if _, _, err = ParseValue(b1); err != nil {
			return
		}
	}
	return m, size, nil
}

// Empty returns true if bytes are empty or message has no fields.
func (m Message) Empty() bool {
	return len(m.bytes) == 0 || m.table.Len() == 0
}

// Len returns the length of the message bytes.
func (m Message) Len() int {
	return len(m.bytes)
}

// Raw returns the underlying message bytes.
func (m Message) Raw() []byte {
	return m.bytes
}

// Fields

// Fields returns the number of fields in the message.
func (m Message) Fields() int {
	return m.table.Len()
}

// HasField returns true if the message contains a field.
func (m Message) HasField(tag uint16) bool {
	end := m.table.Offset(tag)
	size := m.table.DataSize()
	return end >= 0 && end <= int(size)
}

// Field returns a truncated field data as a value or nil.
func (m Message) Field(tag uint16) Value {
	end := m.table.Offset(tag)
	size := m.table.DataSize()

	if end < 0 || end > int(size) {
		return nil
	}

	b := m.bytes[:end]
	return OpenValue(b)
}

// FieldAt returns field data at an index or nil.
func (m Message) FieldAt(i int) Value {
	end := m.table.OffsetByIndex(i)
	size := m.table.DataSize()

	if end < 0 || end > int(size) {
		return nil
	}

	b := m.bytes[:end]
	return OpenValue(b)
}

// FieldRaw returns a raw untruncated field data by a tag or nil.
// The data is larger than the field value, when not the first field.
func (m Message) FieldRaw(tag uint16) []byte {
	end := m.table.Offset(tag)
	size := m.table.DataSize()

	if end < 0 || end > int(size) {
		return nil
	}

	return m.bytes[:end]
}

// Tags

// TagAt returns a field tag at an index or false.
func (m Message) TagAt(i int) (uint16, bool) {
	field, ok := m.table.Field(i)
	if !ok {
		return 0, false
	}
	return field.Tag, true
}

// Clone

// Message returns a message clone.
func (m Message) Clone() Message {
	b := make([]byte, len(m.bytes))
	copy(b, m.bytes)
	return OpenMessage(b)
}

// CloneTo clones a message into a slice, allocates a new slice when needed.
func (m Message) CloneTo(b []byte) Message {
	ln := len(m.bytes)
	if cap(b) < ln {
		b = make([]byte, ln)
	}
	b = b[:ln]

	copy(b, m.bytes)
	return OpenMessage(b)
}

// CloneToArena clones a message into an arena.
func (m Message) CloneToArena(a alloc.Arena) Message {
	n := len(m.bytes)
	buf := alloc.Bytes(a, n)
	copy(buf, m.bytes)
	return OpenMessage(buf)
}

// CloneToBuffer clones a message into a buffer, grows the buffer.
func (m Message) CloneToBuffer(buf buffer.Buffer) Message {
	ln := len(m.bytes)
	b := buf.Grow(ln)
	copy(b, m.bytes)
	return OpenMessage(b)
}

// Types

// Bool decodes and returns a bool or false.
func (m Message) Bool(tag uint16) bool {
	b := m.field(tag)
	v, _, _ := decode.DecodeBool(b)
	return v
}

// Byte decodes and returns a byte or 0.
func (m Message) Byte(tag uint16) byte {
	b := m.field(tag)
	v, _, _ := decode.DecodeByte(b)
	return v
}

// Int

// Int16 decodes and returns an int16 or 0.
func (m Message) Int16(tag uint16) int16 {
	b := m.field(tag)
	v, _, _ := decode.DecodeInt16(b)
	return v
}

// Int32 decodes and returns an int32 or 0.
func (m Message) Int32(tag uint16) int32 {
	b := m.field(tag)
	v, _, _ := decode.DecodeInt32(b)
	return v
}

// Int64 decodes and returns an int64 or 0.
func (m Message) Int64(tag uint16) int64 {
	b := m.field(tag)
	v, _, _ := decode.DecodeInt64(b)
	return v
}

// Uint

// Uint16 decodes and returns a uint16 or 0.
func (m Message) Uint16(tag uint16) uint16 {
	b := m.field(tag)
	v, _, _ := decode.DecodeUint16(b)
	return v
}

// Uint32 decodes and returns a uint32 or 0.
func (m Message) Uint32(tag uint16) uint32 {
	b := m.field(tag)
	v, _, _ := decode.DecodeUint32(b)
	return v
}

// Uint64 decodes and returns a uint64 or 0.
func (m Message) Uint64(tag uint16) uint64 {
	b := m.field(tag)
	v, _, _ := decode.DecodeUint64(b)
	return v
}

// Float

// Float32 decodes and returns a float32 or 0.
func (m Message) Float32(tag uint16) float32 {
	b := m.field(tag)
	v, _, _ := decode.DecodeFloat32(b)
	return v
}

// Float64 decodes and returns a float64 or 0.
func (m Message) Float64(tag uint16) float64 {
	b := m.field(tag)
	v, _, _ := decode.DecodeFloat64(b)
	return v
}

// Bin

// Bin64 decodes and returns a bin64 or a zero value.
func (m Message) Bin64(tag uint16) bin.Bin64 {
	b := m.field(tag)
	v, _, _ := decode.DecodeBin64(b)
	return v
}

// Bin128 decodes and returns a bin128 or a zero value.
func (m Message) Bin128(tag uint16) bin.Bin128 {
	b := m.field(tag)
	v, _, _ := decode.DecodeBin128(b)
	return v
}

// Bin256 decodes and returns a bin256 or a zero value.
func (m Message) Bin256(tag uint16) bin.Bin256 {
	b := m.field(tag)
	v, _, _ := decode.DecodeBin256(b)
	return v
}

// Bytes/string

// Bytes decodes and returns bytes or nil.
func (m Message) Bytes(tag uint16) format.Bytes {
	b := m.field(tag)
	p, _, _ := decode.DecodeBytes(b)
	return p
}

// String decodes and returns a string or an empty string.
func (m Message) String(tag uint16) format.String {
	b := m.field(tag)
	p, _, _ := decode.DecodeString(b)
	return format.String(p)
}

// List/message

// List decodes and returns a list or an empty list.
func (m Message) List(tag uint16) List {
	b := m.field(tag)
	return OpenList(b)
}

// Message decodes and returns a message or an empty message.
func (m Message) Message(tag uint16) Message {
	b := m.field(tag)
	return OpenMessage(b)
}

// internal

func (m Message) field(tag uint16) []byte {
	end := m.table.Offset(tag)
	size := m.table.DataSize()

	if end < 0 || end > int(size) {
		return nil
	}

	return m.bytes[:end]
}

func (m Message) fieldAt(i int) []byte {
	end := m.table.OffsetByIndex(i)
	size := m.table.DataSize()

	if end < 0 || end > int(size) {
		return nil
	}

	return m.bytes[:end]
}
