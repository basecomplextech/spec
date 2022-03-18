package spec

import (
	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
)

type Message struct {
	meta messageMeta
	data []byte
}

// GetMessage decodes and returns a message without recursive validation, or an empty message on error.
func GetMessage(b []byte) Message {
	meta, n, err := decodeMessageMeta(b)
	if err != nil {
		return Message{}
	}

	data := b[len(b)-n:]
	return Message{
		meta: meta,
		data: data,
	}
}

// DecodeMessage decodes, recursively vaildates and returns a message.
func DecodeMessage(b []byte) (Message, int, error) {
	meta, n, err := decodeMessageMeta(b)
	if err != nil {
		return Message{}, n, err
	}
	data := b[len(b)-n:]

	m := Message{
		meta: meta,
		data: data,
	}
	if err := m.Validate(); err != nil {
		return Message{}, n, err
	}
	return m, n, nil
}

// Data returns the exact message data.
func (m Message) Data() []byte {
	return m.data
}

// Count returns the number of fields in the message.
func (m Message) Count() int {
	return m.meta.count()
}

// Field returns field data by a tag or nil.
func (m Message) Field(tag uint16) []byte {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return nil
	case end > int(m.meta.body):
		return nil
	}
	return m.data[:end]
}

// FieldByIndex returns field data by an index or nil.
func (m Message) FieldByIndex(i int) []byte {
	end := m.meta.offsetByIndex(i)
	switch {
	case end < 0:
		return nil
	case end > int(m.meta.body):
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
		if _, _, err := ReadValue(data); err != nil {
			return err
		}
	}
	return nil
}

// Direct access

func (m Message) Bool(tag uint16) bool {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return false
	case end > int(m.meta.body):
		return false
	}

	b := m.data[:end]
	v, _, _ := DecodeBool(b)
	return v
}

func (m Message) Byte(tag uint16) byte {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.meta.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := DecodeByte(b)
	return v
}

func (m Message) Int32(tag uint16) int32 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.meta.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := DecodeInt32(b)
	return v
}

func (m Message) Int64(tag uint16) int64 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.meta.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := DecodeInt64(b)
	return v
}

func (m Message) Uint32(tag uint16) uint32 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.meta.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := DecodeUint32(b)
	return v
}

func (m Message) Uint64(tag uint16) uint64 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.meta.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := DecodeUint64(b)
	return v
}

func (m Message) U128(tag uint16) u128.U128 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return u128.U128{}
	case end > int(m.meta.body):
		return u128.U128{}
	}

	b := m.data[:end]
	v, _, _ := DecodeU128(b)
	return v
}

func (m Message) U256(tag uint16) u256.U256 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return u256.U256{}
	case end > int(m.meta.body):
		return u256.U256{}
	}

	b := m.data[:end]
	v, _, _ := DecodeU256(b)
	return v
}

func (m Message) Float32(tag uint16) float32 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.meta.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := DecodeFloat32(b)
	return v
}

func (m Message) Float64(tag uint16) float64 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.meta.body):
		return 0
	}

	b := m.data[:end]
	v, _, _ := DecodeFloat64(b)
	return v
}

func (m Message) Bytes(tag uint16) []byte {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return nil
	case end > int(m.meta.body):
		return nil
	}

	b := m.data[:end]
	v, _, _ := DecodeBytes(b)
	return v
}

func (m Message) String(tag uint16) string {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return ""
	case end > int(m.meta.body):
		return ""
	}

	b := m.data[:end]
	v, _, _ := DecodeString(b)
	return v
}

func (m Message) Message(tag uint16) Message {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return Message{}
	case end > int(m.meta.body):
		return Message{}
	}

	b := m.data[:end]
	return GetMessage(b)
}
