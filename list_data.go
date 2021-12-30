package spec

type ListData struct {
	l list
}

// NewListData reads and returns list data, but does not validate its elements.
func NewListData(b []byte) (ListData, error) {
	l, err := readList(b)
	if err != nil {
		return ListData{}, err
	}
	return ListData{l}, nil
}

// ReadListData reads and returns list data, and recursively validates its elements.
func ReadListData(b []byte) (ListData, error) {
	l, err := readList(b)
	if err != nil {
		return ListData{}, err
	}

	d := ListData{l}
	if err := d.validate(); err != nil {
		return ListData{}, err
	}
	return d, nil
}

// Data returns the exact list bytes.
func (d ListData) Data() []byte {
	return d.l.data
}

// Reflect access

// Element returns a list element data.
func (d ListData) Element(i int) Data {
	return d.l.element(i)
}

// Len returns the number of elements in the list.
func (d ListData) Len() int {
	return d.l.len()
}

// Direct access

func (d ListData) Bool(i int) bool {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return false
	case end > int(d.l.body):
		return false
	}

	b := d.l.data[start:end]
	v, _ := readBool(b)
	return v
}

func (d ListData) Int8(i int) int8 {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(d.l.body):
		return 0
	}

	b := d.l.data[start:end]
	v, _ := readInt8(b)
	return v
}

func (d ListData) Int16(i int) int16 {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(d.l.body):
		return 0
	}

	b := d.l.data[start:end]
	v, _ := readInt16(b)
	return v
}

func (d ListData) Int32(i int) int32 {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(d.l.body):
		return 0
	}

	b := d.l.data[start:end]
	v, _ := readInt32(b)
	return v
}

func (d ListData) Int64(i int) int64 {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(d.l.body):
		return 0
	}

	b := d.l.data[start:end]
	v, _ := readInt64(b)
	return v
}

func (d ListData) Uint8(i int) uint8 {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(d.l.body):
		return 0
	}

	b := d.l.data[start:end]
	v, _ := readUint8(b)
	return v
}

func (d ListData) Uint16(i int) uint16 {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(d.l.body):
		return 0
	}

	b := d.l.data[start:end]
	v, _ := readUint16(b)
	return v
}

func (d ListData) Uint32(i int) uint32 {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(d.l.body):
		return 0
	}

	b := d.l.data[start:end]
	v, _ := readUint32(b)
	return v
}

func (d ListData) Uint64(i int) uint64 {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(d.l.body):
		return 0
	}

	b := d.l.data[start:end]
	v, _ := readUint64(b)
	return v
}

func (d ListData) Float32(i int) float32 {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(d.l.body):
		return 0
	}

	b := d.l.data[start:end]
	v, _ := readFloat32(b)
	return v
}

func (d ListData) Float64(i int) float64 {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(d.l.body):
		return 0
	}

	b := d.l.data[start:end]
	v, _ := readFloat64(b)
	return v
}

func (d ListData) Bytes(i int) []byte {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return nil
	case end > int(d.l.body):
		return nil
	}

	b := d.l.data[start:end]
	v, _ := readBytes(b)
	return v
}

func (d ListData) String(i int) string {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return ""
	case end > int(d.l.body):
		return ""
	}

	b := d.l.data[start:end]
	v, _ := readString(b)
	return v
}

func (d ListData) List(i int) ListData {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return ListData{}
	case end > int(d.l.body):
		return ListData{}
	}

	b := d.l.data[start:end]
	v, _ := NewListData(b)
	return v
}

func (d ListData) Message(i int) MessageData {
	start, end := d.l.table.offset(d.l.big, i)
	switch {
	case start < 0:
		return MessageData{}
	case end > int(d.l.body):
		return MessageData{}
	}

	b := d.l.data[start:end]
	v, _ := NewMessageData(b)
	return v
}

// private

// validate recursively validates the list.
func (d ListData) validate() error {
	n := d.l.len()

	for i := 0; i < n; i++ {
		data := d.l.element(i)
		if len(data) == 0 {
			continue
		}
		if _, err := ReadData(data); err != nil {
			return err
		}
	}
	return nil
}
