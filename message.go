package spec

type Message struct {
	message
}

// GetMessage parses and returns a message, but does not validate it.
func GetMessage(b []byte) (Message, error) {
	m, err := readMessage(b)
	if err != nil {
		return Message{}, err
	}
	return Message{m}, nil
}

// ReadMessage reads, recursively validates and returns a message.
func ReadMessage(b []byte) (Message, error) {
	m, err := readMessage(b)
	if err != nil {
		return Message{}, err
	}
	msg := Message{m}
	if err := msg.Validate(); err != nil {
		return Message{}, err
	}
	return msg, nil
}

// Data returns the exact message bytes.
func (m Message) Data() []byte {
	return m.data
}

// Validate recursively validates the message.
func (m Message) Validate() error {
	n := m.Len()

	for i := 0; i < n; i++ {
		data := m.FieldByIndex(i)
		if len(data) == 0 {
			continue
		}
		if _, err := ReadValue(data); err != nil {
			return err
		}
	}
	return nil
}

// Element returns a field data by a tag or nil.
func (m Message) Element(tag uint16) []byte {
	return m.element(tag)
}

// Field returns a field data by a tag or nil.
func (m Message) Field(tag uint16) []byte {
	return m.field(tag)
}

// FieldByIndex returns a field data by an index or nil.
func (m Message) FieldByIndex(i int) []byte {
	return m.fieldByIndex(i)
}

// Len returns the number of fields in the message.
func (m Message) Len() int {
	return m.len()
}

// Get

func (m Message) Bool(tag uint16) bool {
	end := m.table.offset(m.big, tag)
	switch {
	case end < 0:
		return false
	case end > int(m.body):
		return false
	}

	b := m.data[:end]
	v, _ := readBool(b)
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
	v, _ := readInt8(b)
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
	v, _ := readInt16(b)
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
	v, _ := readInt32(b)
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
	v, _ := readInt64(b)
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
	v, _ := readUint8(b)
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
	v, _ := readUint16(b)
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
	v, _ := readUint32(b)
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
	v, _ := readUint64(b)
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
	v, _ := readFloat32(b)
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
	v, _ := readFloat64(b)
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
	v, _ := readBytes(b)
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
	v, _ := readString(b)
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
	v, _ := GetList(b)
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
	v, _ := GetMessage(b)
	return v
}
