package spec

import (
	"github.com/complex1tech/baselibrary/types"
	"github.com/complex1tech/spec/go/encoding"
)

type Message struct {
	meta  encoding.MessageMeta
	bytes []byte
}

// GetMessage decodes and returns a message without recursive validation, or an empty message on error.
func GetMessage(b []byte) Message {
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

// DecodeMessage decodes, recursively vaildates and returns a message.
func DecodeMessage(b []byte) (_ Message, size int, err error) {
	meta, size, err := encoding.DecodeMessageMeta(b)
	if err != nil {
		return
	}
	bytes := b[len(b)-size:]

	m := Message{
		meta:  meta,
		bytes: bytes,
	}

	ln := m.Len()
	for i := 0; i < ln; i++ {
		field := m.FieldByIndex(i)
		if len(field) == 0 {
			continue
		}
		if _, _, err = DecodeValue(field); err != nil {
			return
		}
	}
	return m, size, nil
}

// Message returns a message clone.
func (m Message) Clone() Message {
	b := make([]byte, len(m.bytes))
	copy(b, m.bytes)
	return GetMessage(b)
}

// CloneTo clones a message into a byte slice.
func (m Message) CloneTo(b []byte) Message {
	ln := len(m.bytes)
	if cap(b) < ln {
		b = make([]byte, ln)
	}
	b = b[:ln]

	copy(b, m.bytes)
	return GetMessage(b)
}

// Len returns the number of fields in the message.
func (m Message) Len() int {
	return m.meta.Len()
}

// Bytes returns the exact message bytes.
func (m Message) Bytes() []byte {
	return m.bytes
}

// Empty returns true if the message is backed byte an empty byte slice or has no fields.
func (m Message) Empty() bool {
	return len(m.bytes) == 0 || m.meta.Len() == 0
}

// Field returns field data by a tag or nil.
func (m Message) Field(tag uint16) []byte {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return nil
	case end > int(size):
		return nil
	}
	return m.bytes[:end]
}

// FieldByIndex returns field data by an index or nil.
func (m Message) FieldByIndex(i int) []byte {
	end := m.meta.OffsetByIndex(i)
	size := m.meta.Data()

	switch {
	case end < 0:
		return nil
	case end > int(size):
		return nil
	}
	return m.bytes[:end]
}

// HasField returns true if the message contains a field.
func (m Message) HasField(tag uint16) bool {
	end := m.meta.Offset(tag)
	size := m.meta.Data()
	return end >= 0 && end <= int(size)
}

// TagByIndex returns a field tag by index or false.
func (m Message) TagByIndex(i int) (uint16, bool) {
	field, ok := m.meta.Field(i)
	if !ok {
		return 0, false
	}
	return field.Tag, true
}

// Direct access

func (m Message) GetBool(tag uint16) bool {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return false
	case end > int(size):
		return false
	}

	b := m.bytes[:end]
	v, _, _ := encoding.DecodeBool(b)
	return v
}

func (m Message) GetByte(tag uint16) byte {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return 0
	case end > int(size):
		return 0
	}

	b := m.bytes[:end]
	v, _, _ := encoding.DecodeByte(b)
	return v
}

func (m Message) GetInt32(tag uint16) int32 {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return 0
	case end > int(size):
		return 0
	}

	b := m.bytes[:end]
	v, _, _ := encoding.DecodeInt32(b)
	return v
}

func (m Message) GetInt64(tag uint16) int64 {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return 0
	case end > int(size):
		return 0
	}

	b := m.bytes[:end]
	v, _, _ := encoding.DecodeInt64(b)
	return v
}

func (m Message) GetUint32(tag uint16) uint32 {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return 0
	case end > int(size):
		return 0
	}

	b := m.bytes[:end]
	v, _, _ := encoding.DecodeUint32(b)
	return v
}

func (m Message) GetUint64(tag uint16) uint64 {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return 0
	case end > int(size):
		return 0
	}

	b := m.bytes[:end]
	v, _, _ := encoding.DecodeUint64(b)
	return v
}

func (m Message) GetBin64(tag uint16) types.Bin64 {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return types.Bin64{}
	case end > int(size):
		return types.Bin64{}
	}

	b := m.bytes[:end]
	v, _, _ := encoding.DecodeBin64(b)
	return v
}

func (m Message) GetBin128(tag uint16) types.Bin128 {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return types.Bin128{}
	case end > int(size):
		return types.Bin128{}
	}

	b := m.bytes[:end]
	v, _, _ := encoding.DecodeBin128(b)
	return v
}

func (m Message) GetBin256(tag uint16) types.Bin256 {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return types.Bin256{}
	case end > int(size):
		return types.Bin256{}
	}

	b := m.bytes[:end]
	v, _, _ := encoding.DecodeBin256(b)
	return v
}

func (m Message) GetFloat32(tag uint16) float32 {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return 0
	case end > int(size):
		return 0
	}

	b := m.bytes[:end]
	v, _, _ := encoding.DecodeFloat32(b)
	return v
}

func (m Message) GetFloat64(tag uint16) float64 {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return 0
	case end > int(size):
		return 0
	}

	b := m.bytes[:end]
	v, _, _ := encoding.DecodeFloat64(b)
	return v
}

func (m Message) GetBytes(tag uint16) []byte {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return nil
	case end > int(size):
		return nil
	}

	b := m.bytes[:end]
	v, _, _ := encoding.DecodeBytes(b)
	return v
}

func (m Message) GetString(tag uint16) string {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return ""
	case end > int(size):
		return ""
	}

	b := m.bytes[:end]
	v, _, _ := encoding.DecodeString(b)
	return v
}

func (m Message) GetMessage(tag uint16) Message {
	end := m.meta.Offset(tag)
	size := m.meta.Data()

	switch {
	case end < 0:
		return Message{}
	case end > int(size):
		return Message{}
	}

	b := m.bytes[:end]
	return GetMessage(b)
}
