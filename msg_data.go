package spec

type MessageData struct {
	m message
}

// NewMessageData reads and returns message data, but does not validate its fields.
func NewMessageData(b []byte) (MessageData, error) {
	m, err := readMessage(b)
	if err != nil {
		return MessageData{}, err
	}
	return MessageData{m}, nil
}

// ReadMessageData reads and returns message data, and recursively validates its fields.
func ReadMessageData(b []byte) (MessageData, error) {
	m, err := readMessage(b)
	if err != nil {
		return MessageData{}, err
	}

	d := MessageData{m}
	if err := d.validate(); err != nil {
		return MessageData{}, err
	}
	return d, nil
}

// Data returns the exact message bytes.
func (d MessageData) Data() []byte {
	return d.m.data
}

// Reflect access

// Element returns a field data by a tag or nil.
// The method is an alias for field.
func (d MessageData) Element(tag uint16) Data {
	return d.m.element(tag)
}

// Field returns a field data by a tag or nil.
func (d MessageData) Field(tag uint16) Data {
	return d.m.field(tag)
}

// FieldByIndex returns a field data by an index or nil.
func (d MessageData) FieldByIndex(i int) Data {
	return d.m.fieldByIndex(i)
}

// Len returns the number of fields in the message.
func (d MessageData) Len() int {
	return d.m.len()
}

// Direct access

func (d MessageData) Bool(tag uint16) bool {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return false
	case end > int(d.m.body):
		return false
	}

	b := d.m.data[:end]
	v, _ := readBool(b)
	return v
}

func (d MessageData) Int8(tag uint16) int8 {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.m.body):
		return 0
	}

	b := d.m.data[:end]
	v, _ := readInt8(b)
	return v
}

func (d MessageData) Int16(tag uint16) int16 {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.m.body):
		return 0
	}

	b := d.m.data[:end]
	v, _ := readInt16(b)
	return v
}

func (d MessageData) Int32(tag uint16) int32 {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.m.body):
		return 0
	}

	b := d.m.data[:end]
	v, _ := readInt32(b)
	return v
}

func (d MessageData) Int64(tag uint16) int64 {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.m.body):
		return 0
	}

	b := d.m.data[:end]
	v, _ := readInt64(b)
	return v
}

func (d MessageData) Uint8(tag uint16) uint8 {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.m.body):
		return 0
	}

	b := d.m.data[:end]
	v, _ := readUint8(b)
	return v
}

func (d MessageData) Uint16(tag uint16) uint16 {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.m.body):
		return 0
	}

	b := d.m.data[:end]
	v, _ := readUint16(b)
	return v
}

func (d MessageData) Uint32(tag uint16) uint32 {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.m.body):
		return 0
	}

	b := d.m.data[:end]
	v, _ := readUint32(b)
	return v
}

func (d MessageData) Uint64(tag uint16) uint64 {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.m.body):
		return 0
	}

	b := d.m.data[:end]
	v, _ := readUint64(b)
	return v
}

func (d MessageData) Float32(tag uint16) float32 {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.m.body):
		return 0
	}

	b := d.m.data[:end]
	v, _ := readFloat32(b)
	return v
}

func (d MessageData) Float64(tag uint16) float64 {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.m.body):
		return 0
	}

	b := d.m.data[:end]
	v, _ := readFloat64(b)
	return v
}

func (d MessageData) Bytes(tag uint16) []byte {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return nil
	case end > int(d.m.body):
		return nil
	}

	b := d.m.data[:end]
	v, _ := readBytes(b)
	return v
}

func (d MessageData) String(tag uint16) string {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return ""
	case end > int(d.m.body):
		return ""
	}

	b := d.m.data[:end]
	v, _ := readString(b)
	return v
}

func (d MessageData) List(tag uint16) ListData {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return ListData{}
	case end > int(d.m.body):
		return ListData{}
	}

	b := d.m.data[:end]
	v, _ := NewListData(b)
	return v
}

func (d MessageData) Message(tag uint16) MessageData {
	end := d.m.table.offset(d.m.big, tag)
	switch {
	case end < 0:
		return MessageData{}
	case end > int(d.m.body):
		return MessageData{}
	}

	b := d.m.data[:end]
	v, _ := NewMessageData(b)
	return v
}

// private

// validate recursively validates the message.
func (d MessageData) validate() error {
	n := d.m.len()

	for i := 0; i < n; i++ {
		data := d.m.fieldByIndex(i)
		if len(data) == 0 {
			continue
		}
		if _, err := ReadData(data); err != nil {
			return err
		}
	}
	return nil
}
