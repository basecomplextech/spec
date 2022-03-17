package spec

import (
	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
)

type Message struct {
	data  []byte
	table messageTable

	body uint32 // body size
	big  bool   // big/small table format
}

// NewMessage reads and returns a message, but does not validate its fields.
func NewMessage(b []byte) (Message, error) {
	m, _, err := readMessage(b)
	return m, err
}

// ReadMessage reads, recursively vaildates and returns a message.
func ReadMessage(b []byte) (Message, error) {
	m, _, err := readMessage(b)
	if err != nil {
		return Message{}, err
	}

	if err := m.Validate(); err != nil {
		return Message{}, err
	}
	return m, nil
}

// Data returns the exact message data.
func (m Message) Data() []byte {
	return m.data
}

// Count returns the number of fields in the message.
func (m Message) Count() int {
	return m.table.count(m.big)
}

// Field returns field data by a tag or nil.
func (m Message) Field(tag uint16) []byte {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return nil
	case end > int(m.body):
		return nil
	}
	return m.data[:end]
}

// FieldByIndex returns field data by an index or nil.
func (m Message) FieldByIndex(i int) []byte {
	end := m.table.offsetByIndex(m.big, i)
	switch {
	case end < 0:
		return nil
	case end > int(m.body):
		return nil
	}
	return m.data[:end]
}

// Validate recursively validates the message.
func (m Message) Validate() error {
	n := m.Count()

	for i := 0; i < n; i++ {
		data := m.FieldByIndex(i)
		if len(data) == 0 {
			continue
		}
		if _, err := ReadData(data); err != nil {
			return err
		}
	}
	return nil
}

// Direct access

func (m Message) Bool(tag uint16) bool {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return false
	case end > int(m.body):
		return false
	}

	b := m.data[:end]
	v, _, _ := readBool(b)
	return v
}

func (m Message) Int8(tag uint16) int8 {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := readInt8(b)
	return v
}

func (m Message) Int16(tag uint16) int16 {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := readInt16(b)
	return v
}

func (m Message) Int32(tag uint16) int32 {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := readInt32(b)
	return v
}

func (m Message) Int64(tag uint16) int64 {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := readInt64(b)
	return v
}

func (m Message) Uint8(tag uint16) uint8 {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := readUint8(b)
	return v
}

func (m Message) Uint16(tag uint16) uint16 {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := readUint16(b)
	return v
}

func (m Message) Uint32(tag uint16) uint32 {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := readUint32(b)
	return v
}

func (m Message) Uint64(tag uint16) uint64 {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := readUint64(b)
	return v
}

func (m Message) U128(tag uint16) u128.U128 {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return u128.U128{}
	case end > int(m.body):
		return u128.U128{}
	}

	b := m.data[:end]
	v, _, _ := readU128(b)
	return v
}

func (m Message) U256(tag uint16) u256.U256 {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return u256.U256{}
	case end > int(m.body):
		return u256.U256{}
	}

	b := m.data[:end]
	v, _, _ := readU256(b)
	return v
}

func (m Message) Float32(tag uint16) float32 {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := readFloat32(b)
	return v
}

func (m Message) Float64(tag uint16) float64 {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := readFloat64(b)
	return v
}

func (m Message) Bytes(tag uint16) []byte {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return nil
	case end > int(m.body):
		return nil
	}

	b := m.data[:end]
	v, _, _ := readBytes(b)
	return v
}

func (m Message) String(tag uint16) string {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return ""
	case end > int(m.body):
		return ""
	}

	b := m.data[:end]
	v, _, _ := readString(b)
	return v
}

func (m Message) List(tag uint16) List {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return List{}
	case end > int(m.body):
		return List{}
	}

	b := m.data[:end]
	v, _ := NewList(b)
	return v
}

func (m Message) Message(tag uint16) Message {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return Message{}
	case end > int(m.body):
		return Message{}
	}

	b := m.data[:end]
	v, _ := NewMessage(b)
	return v
}
