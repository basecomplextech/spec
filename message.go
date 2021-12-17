package spec

type Message struct {
	buffer []byte
	table  messageTable
}

func ReadMessage(buf []byte) Message {
	type_, b := readType(buf)
	if type_ != TypeMessage {
		return Message{}
	}

	tsize, b := readMessageTableSize(b)
	dsize, b := readMessageDataSize(b)
	table, _ := readMessageTable(b, tsize)
	buffer := readMessageBuffer(buf, tsize, dsize) // slice initial buf

	return Message{
		buffer: buffer,
		table:  table,
	}
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
	b := m.buffer[:off]
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

func (m Message) Int8(tag uint16) int8 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.buffer[:off]
	return ReadInt8(b)
}

func (m Message) Int16(tag uint16) int16 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.buffer[:off]
	return ReadInt16(b)
}

func (m Message) Int32(tag uint16) int32 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.buffer[:off]
	return ReadInt32(b)
}

func (m Message) Int64(tag uint16) int64 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.buffer[:off]
	return ReadInt64(b)
}

func (m Message) UInt8(tag uint16) uint8 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.buffer[:off]
	return ReadUInt8(b)
}

func (m Message) UInt16(tag uint16) uint16 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.buffer[:off]
	return ReadUInt16(b)
}

func (m Message) UInt32(tag uint16) uint32 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.buffer[:off]
	return ReadUInt32(b)
}

func (m Message) UInt64(tag uint16) uint64 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.buffer[:off]
	return ReadUInt64(b)
}

func (m Message) Float32(tag uint16) float32 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.buffer[:off]
	return ReadFloat32(b)
}

func (m Message) Float64(tag uint16) float64 {
	off := m.table.offset(tag)
	if off < 0 {
		return 0
	}
	b := m.buffer[:off]
	return ReadFloat64(b)
}

func (m Message) Bytes(tag uint16) []byte {
	off := m.table.offset(tag)
	if off < 0 {
		return nil
	}
	b := m.buffer[:off]
	return ReadBytes(b)
}

func (m Message) String(tag uint16) string {
	off := m.table.offset(tag)
	if off < 0 {
		return ""
	}
	b := m.buffer[:off]
	return ReadString(b)
}

func (m Message) List(tag uint16) List {
	off := m.table.offset(tag)
	if off < 0 {
		return List{}
	}
	b := m.buffer[:off]
	return ReadList(b)
}
