package spec

type MessageData struct {
	message
}

// NewMessageData parses and returns message data, but does not validate it.
func NewMessageData(b []byte) (MessageData, error) {
	d, err := readMessage(b)
	if err != nil {
		return MessageData{}, err
	}
	return MessageData{d}, nil
}

// ReadMessageData reads, recursively validates and returns message data.
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

// validate recursively validates the message.
func (d MessageData) validate() error {
	n := d.len()

	for i := 0; i < n; i++ {
		data := d.fieldByIndex(i)
		if len(data) == 0 {
			continue
		}
		if _, err := ReadData(data); err != nil {
			return err
		}
	}
	return nil
}

// Data returns the exact message bytes.
func (d MessageData) Data() []byte {
	return d.data
}

// Element returns a field data by a tag or nil.
func (d MessageData) Element(tag uint16) Data {
	return d.element(tag)
}

// Field returns a field data by a tag or nil.
func (d MessageData) Field(tag uint16) Data {
	return d.field(tag)
}

// FieldByIndex returns a field data by an index or nil.
func (d MessageData) FieldByIndex(i int) Data {
	return d.fieldByIndex(i)
}

// Len returns the number of fields in the message.
func (d MessageData) Len() int {
	return d.len()
}

// Getters

func (d MessageData) Bool(tag uint16) bool {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return false
	case end > int(d.body):
		return false
	}

	b := d.data[:end]
	v, _ := readBool(b)
	return v
}

func (d MessageData) Int8(tag uint16) int8 {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.body):
		return 0
	}

	b := d.data[:end]
	v, _ := readInt8(b)
	return v
}

func (d MessageData) Int16(tag uint16) int16 {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.body):
		return 0
	}

	b := d.data[:end]
	v, _ := readInt16(b)
	return v
}

func (d MessageData) Int32(tag uint16) int32 {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.body):
		return 0
	}

	b := d.data[:end]
	v, _ := readInt32(b)
	return v
}

func (d MessageData) Int64(tag uint16) int64 {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.body):
		return 0
	}

	b := d.data[:end]
	v, _ := readInt64(b)
	return v
}

func (d MessageData) Uint8(tag uint16) uint8 {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.body):
		return 0
	}

	b := d.data[:end]
	v, _ := readUint8(b)
	return v
}

func (d MessageData) Uint16(tag uint16) uint16 {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.body):
		return 0
	}

	b := d.data[:end]
	v, _ := readUint16(b)
	return v
}

func (d MessageData) Uint32(tag uint16) uint32 {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.body):
		return 0
	}

	b := d.data[:end]
	v, _ := readUint32(b)
	return v
}

func (d MessageData) Uint64(tag uint16) uint64 {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.body):
		return 0
	}

	b := d.data[:end]
	v, _ := readUint64(b)
	return v
}

func (d MessageData) Float32(tag uint16) float32 {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.body):
		return 0
	}

	b := d.data[:end]
	v, _ := readFloat32(b)
	return v
}

func (d MessageData) Float64(tag uint16) float64 {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return 0
	case end > int(d.body):
		return 0
	}

	b := d.data[:end]
	v, _ := readFloat64(b)
	return v
}

func (d MessageData) Bytes(tag uint16) []byte {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return nil
	case end > int(d.body):
		return nil
	}

	b := d.data[:end]
	v, _ := readBytes(b)
	return v
}

func (d MessageData) String(tag uint16) string {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return ""
	case end > int(d.body):
		return ""
	}

	b := d.data[:end]
	v, _ := readString(b)
	return v
}

func (d MessageData) List(tag uint16) ListData {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return ListData{}
	case end > int(d.body):
		return ListData{}
	}

	b := d.data[:end]
	v, _ := NewListData(b)
	return v
}

func (d MessageData) Message(tag uint16) MessageData {
	end := d.table.offset(d.big, tag)
	switch {
	case end < 0:
		return MessageData{}
	case end > int(d.body):
		return MessageData{}
	}

	b := d.data[:end]
	v, _ := NewMessageData(b)
	return v
}
