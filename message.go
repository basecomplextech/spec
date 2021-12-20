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
func (m Message) Field(tag uint16) (d Data) {
	off := m.table.offset(tag)
	if off < 0 {
		return
	}
	b := m.data[:off]
	return ReadData(b)
}

// FieldByIndex returns a field value by an index or an empty value.
func (m Message) FieldByIndex(i int) (d Data, ok bool) {
	f, ok := m.table.field(i)
	if !ok {
		return
	}

	b := m.buffer[:f.offset]
	d = ReadData(b)
	ok = true
	return
}

// Fields returns the number of fields in the message.
func (m Message) Fields() int {
	return m.table.count()
}

// Readers

func (m Message) Bool(tag uint16) bool {
	off := m.table.offset(tag)
	if off < 0 {
		return false
	}
	b := m.data[:off]
	v, _ := ReadBool(b)
	return v
}

func (m Message) Int8(tag uint16) int8 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.data[:off]
	v, _ := ReadInt8(b)
	return v
}

func (m Message) Int16(tag uint16) int16 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.data[:off]
	v, _ := ReadInt16(b)
	return v
}

func (m Message) Int32(tag uint16) int32 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.data[:off]
	v, _ := ReadInt32(b)
	return v
}

func (m Message) Int64(tag uint16) int64 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.data[:off]
	v, _ := ReadInt64(b)
	return v
}

func (m Message) UInt8(tag uint16) uint8 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.data[:off]
	v, _ := ReadUInt8(b)
	return v
}

func (m Message) UInt16(tag uint16) uint16 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.data[:off]
	v, _ := ReadUInt16(b)
	return v
}

func (m Message) UInt32(tag uint16) uint32 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.data[:off]
	v, _ := ReadUInt32(b)
	return v
}

func (m Message) UInt64(tag uint16) uint64 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.data[:off]
	v, _ := ReadUInt64(b)
	return v
}

func (m Message) Float32(tag uint16) float32 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.data[:off]
	v, _ := ReadFloat32(b)
	return v
}

func (m Message) Float64(tag uint16) float64 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.data[:off]
	v, _ := ReadFloat64(b)
	return v
}

func (m Message) Bytes(tag uint16) []byte {
	off := m.table.offset(tag)
	if off < 0 {
		return nil
	}
	b := m.data[:off]
	v, _ := ReadBytes(b)
	return v
}

func (m Message) String(tag uint16) string {
	off := m.table.offset(tag)
	if off < 0 {
		return ""
	}
	b := m.data[:off]
	v, _ := ReadString(b)
	return v
}

func (m Message) List(tag uint16) List {
	off := m.table.offset(tag)
	if off < 0 {
		return List{}
	}
	b := m.data[:off]
	v, _ := ReadList(b)
	return v
}
