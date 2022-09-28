package spec

import "github.com/complex1tech/baselibrary/types"

type Message struct {
	meta  messageMeta
	bytes []byte
}

// GetMessage decodes and returns a message without recursive validation, or an empty message on error.
func GetMessage(b []byte) Message {
	meta, n, err := decodeMessageMeta(b)
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
	meta, size, err := decodeMessageMeta(b)
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
	return m.meta.count()
}

// Bytes returns the exact message bytes.
func (m Message) Bytes() []byte {
	return m.bytes
}

// Empty returns true if the message is backed byte an empty byte slice or has no fields.
func (m Message) Empty() bool {
	return len(m.bytes) == 0 || m.meta.count() == 0
}

// Field returns field data by a tag or nil.
func (m Message) Field(tag uint16) []byte {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return nil
	case end > int(m.meta.data):
		return nil
	}
	return m.bytes[:end]
}

// FieldByIndex returns field data by an index or nil.
func (m Message) FieldByIndex(i int) []byte {
	end := m.meta.offsetByIndex(i)
	switch {
	case end < 0:
		return nil
	case end > int(m.meta.data):
		return nil
	}
	return m.bytes[:end]
}

// HasField returns true if the message contains a field.
func (m Message) HasField(tag uint16) bool {
	end := m.meta.offset(tag)
	return end >= 0 && end <= int(m.meta.data)
}

// TagByIndex returns a field tag by index or false.
func (m Message) TagByIndex(i int) (uint16, bool) {
	field, ok := m.meta.field(i)
	if !ok {
		return 0, false
	}
	return field.tag, true
}

// Direct access

func (m Message) GetBool(tag uint16) bool {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return false
	case end > int(m.meta.data):
		return false
	}

	b := m.bytes[:end]
	v, _, _ := DecodeBool(b)
	return v
}

func (m Message) GetByte(tag uint16) byte {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.meta.data):
		return 0
	}

	b := m.bytes[:end]
	v, _, _ := DecodeByte(b)
	return v
}

func (m Message) GetInt32(tag uint16) int32 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.meta.data):
		return 0
	}

	b := m.bytes[:end]
	v, _, _ := DecodeInt32(b)
	return v
}

func (m Message) GetInt64(tag uint16) int64 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.meta.data):
		return 0
	}

	b := m.bytes[:end]
	v, _, _ := DecodeInt64(b)
	return v
}

func (m Message) GetUint32(tag uint16) uint32 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.meta.data):
		return 0
	}

	b := m.bytes[:end]
	v, _, _ := DecodeUint32(b)
	return v
}

func (m Message) GetUint64(tag uint16) uint64 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.meta.data):
		return 0
	}

	b := m.bytes[:end]
	v, _, _ := DecodeUint64(b)
	return v
}

func (m Message) GetBin128(tag uint16) types.Bin128 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return types.Bin128{}
	case end > int(m.meta.data):
		return types.Bin128{}
	}

	b := m.bytes[:end]
	v, _, _ := DecodeBin128(b)
	return v
}

func (m Message) GetBin256(tag uint16) types.Bin256 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return types.Bin256{}
	case end > int(m.meta.data):
		return types.Bin256{}
	}

	b := m.bytes[:end]
	v, _, _ := DecodeBin256(b)
	return v
}

func (m Message) GetFloat32(tag uint16) float32 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.meta.data):
		return 0
	}

	b := m.bytes[:end]
	v, _, _ := DecodeFloat32(b)
	return v
}

func (m Message) GetFloat64(tag uint16) float64 {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return 0
	case end > int(m.meta.data):
		return 0
	}

	b := m.bytes[:end]
	v, _, _ := DecodeFloat64(b)
	return v
}

func (m Message) GetBytes(tag uint16) []byte {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return nil
	case end > int(m.meta.data):
		return nil
	}

	b := m.bytes[:end]
	v, _, _ := DecodeBytes(b)
	return v
}

func (m Message) GetString(tag uint16) string {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return ""
	case end > int(m.meta.data):
		return ""
	}

	b := m.bytes[:end]
	v, _, _ := DecodeString(b)
	return v
}

func (m Message) GetMessage(tag uint16) Message {
	end := m.meta.offset(tag)
	switch {
	case end < 0:
		return Message{}
	case end > int(m.meta.data):
		return Message{}
	}

	b := m.bytes[:end]
	return GetMessage(b)
}
