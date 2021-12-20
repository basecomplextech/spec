package spec

type Message struct {
	buffer []byte
	table  messageTable
	data   []byte
}

// Data returns the exact message bytes.
func (m Message) Data() []byte {
	return m.buffer
}

// Field returns a field value by a tag or an empty value.
func (m Message) Field(tag uint16) []byte {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return nil
	case off > len(m.data):
		return nil
	}
	return m.data[:off]
}

// FieldByIndex returns a field value by an index or an empty value.
func (m Message) FieldByIndex(i int) (Value, bool) {
	f, ok := m.table.field(i)
	switch {
	case !ok:
		return Value{}, false
	case int(f.offset) > len(m.data):
		return Value{}, false
	}

	b := m.buffer[:f.offset]
	v := ReadValue(b)
	return v, true
}

// Fields returns the number of fields in the message.
func (m Message) Fields() int {
	return m.table.count()
}

// Getters

func (m Message) Bool(tag uint16) bool {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return false
	case off > len(m.data):
		return false
	}

	b := m.data[:off]
	v, _ := ReadBool(b)
	return v
}

func (m Message) Int8(tag uint16) int8 {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return 0
	case off > len(m.data):
		return 0
	}

	b := m.data[:off]
	v, _ := ReadInt8(b)
	return v
}

func (m Message) Int16(tag uint16) int16 {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return 0
	case off > len(m.data):
		return 0
	}

	b := m.data[:off]
	v, _ := ReadInt16(b)
	return v
}

func (m Message) Int32(tag uint16) int32 {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return 0
	case off > len(m.data):
		return 0
	}

	b := m.data[:off]
	v, _ := ReadInt32(b)
	return v
}

func (m Message) Int64(tag uint16) int64 {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return 0
	case off > len(m.data):
		return 0
	}

	b := m.data[:off]
	v, _ := ReadInt64(b)
	return v
}

func (m Message) UInt8(tag uint16) uint8 {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return 0
	case off > len(m.data):
		return 0
	}

	b := m.data[:off]
	v, _ := ReadUInt8(b)
	return v
}

func (m Message) UInt16(tag uint16) uint16 {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return 0
	case off > len(m.data):
		return 0
	}

	b := m.data[:off]
	v, _ := ReadUInt16(b)
	return v
}

func (m Message) UInt32(tag uint16) uint32 {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return 0
	case off > len(m.data):
		return 0
	}

	b := m.data[:off]
	v, _ := ReadUInt32(b)
	return v
}

func (m Message) UInt64(tag uint16) uint64 {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return 0
	case off > len(m.data):
		return 0
	}

	b := m.data[:off]
	v, _ := ReadUInt64(b)
	return v
}

func (m Message) Float32(tag uint16) float32 {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return 0
	case off > len(m.data):
		return 0
	}

	b := m.data[:off]
	v, _ := ReadFloat32(b)
	return v
}

func (m Message) Float64(tag uint16) float64 {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return 0
	case off > len(m.data):
		return 0
	}

	b := m.data[:off]
	v, _ := ReadFloat64(b)
	return v
}

func (m Message) Bytes(tag uint16) []byte {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return nil
	case off > len(m.data):
		return nil
	}

	b := m.data[:off]
	v, _ := ReadBytes(b)
	return v
}

func (m Message) String(tag uint16) string {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return ""
	case off > len(m.data):
		return ""
	}

	b := m.data[:off]
	v, _ := ReadString(b)
	return v
}

func (m Message) List(tag uint16) List {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return List{}
	case off > len(m.data):
		return List{}
	}

	b := m.data[:off]
	v, _ := ReadList(b)
	return v
}

func (m Message) Message(tag uint16) Message {
	off := m.table.offset(tag)
	switch {
	case off < 0:
		return Message{}
	case off > len(m.data):
		return Message{}
	}

	b := m.data[:off]
	v, _ := ReadMessage(b)
	return v
}
