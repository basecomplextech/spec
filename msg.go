package spec

import (
	"github.com/basecomplextech/spec/encoding"
)

// Message is a raw message.
type Message struct {
	meta  encoding.MessageMeta
	bytes []byte
}

// NewMessage returns a new message from bytes or an empty message when not a message.
func NewMessage(b []byte) Message {
	meta, n, err := encoding.DecodeMessageMeta(b)
	if err != nil {
		return Message{}
	}
	bytes := b[len(b)-n:]

	return Message{
		meta:  meta,
		bytes: bytes,
	}
}

// NewMessageErr returns a new message from bytes or an error when not a message.
func NewMessageErr(b []byte) (Message, error) {
	meta, n, err := encoding.DecodeMessageMeta(b)
	if err != nil {
		return Message{}, err
	}
	bytes := b[len(b)-n:]

	m := Message{
		meta:  meta,
		bytes: bytes,
	}
	return m, nil
}

// ParseMessage recursively parses and returns a message.
func ParseMessage(b []byte) (_ Message, size int, err error) {
	meta, size, err := encoding.DecodeMessageMeta(b)
	if err != nil {
		return Message{}, 0, err
	}
	bytes := b[len(b)-size:]

	m := Message{
		meta:  meta,
		bytes: bytes,
	}

	ln := m.Len()
	for i := 0; i < ln; i++ {
		b1 := m.FieldByIndex(i)
		if len(b1) == 0 {
			continue
		}

		if _, _, err = ParseValue(b1); err != nil {
			return
		}
	}
	return m, size, nil
}

// Len returns the number of fields in the message.
func (m Message) Len() int {
	return m.meta.Len()
}

// Empty returns true if bytes are empty or message has no fields.
func (m Message) Empty() bool {
	return len(m.bytes) == 0 || m.meta.Len() == 0
}

// Raw returns the underlying message bytes.
func (m Message) Raw() []byte {
	return m.bytes
}

// Fields

// HasField returns true if the message contains a field.
func (m Message) HasField(tag uint16) bool {
	end := m.meta.Offset(tag)
	size := m.meta.DataSize()
	return end >= 0 && end <= int(size)
}

// Field returns field data by a tag or nil.
func (m Message) Field(tag uint16) Value {
	end := m.meta.Offset(tag)
	size := m.meta.DataSize()

	switch {
	case end < 0:
		return nil
	case end > int(size):
		return nil
	}

	b := m.bytes[:end]
	return Value(b)
}

// FieldBytes returns field data by a tag or nil.
func (m Message) FieldBytes(tag uint16) []byte {
	end := m.meta.Offset(tag)
	size := m.meta.DataSize()

	switch {
	case end < 0:
		return nil
	case end > int(size):
		return nil
	}

	return m.bytes[:end]
}

// FieldByIndex returns field data by an index or nil.
func (m Message) FieldByIndex(i int) Value {
	end := m.meta.OffsetByIndex(i)
	size := m.meta.DataSize()

	switch {
	case end < 0:
		return nil
	case end > int(size):
		return nil
	}

	b := m.bytes[:end]
	return NewValue(b)
}

// Tags

// TagByIndex returns a field tag by index or false.
func (m Message) TagByIndex(i int) (uint16, bool) {
	field, ok := m.meta.Field(i)
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
	return NewMessage(b)
}

// CloneTo clones a message into a slice.
func (m Message) CloneTo(b []byte) Message {
	ln := len(m.bytes)
	if cap(b) < ln {
		b = make([]byte, ln)
	}
	b = b[:ln]

	copy(b, m.bytes)
	return NewMessage(b)
}
