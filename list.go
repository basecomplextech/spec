package spec

type List struct {
	list
}

// GetList parses and returns a list, but does not validate it.
func GetList(b []byte) (List, error) {
	l, err := readList(b)
	if err != nil {
		return List{}, err
	}
	return List{l}, nil
}

// ReadList reads, recursively validates and returns a list.
func ReadList(b []byte) (List, error) {
	l, err := readList(b)
	if err != nil {
		return List{}, err
	}

	list := List{l}
	if err := list.Validate(); err != nil {
		return List{}, err
	}
	return list, nil
}

// Validate recursively validates the list.
func (l List) Validate() error {
	n := l.Len()

	for i := 0; i < n; i++ {
		data := l.Element(i)
		if len(data) == 0 {
			continue
		}
		if _, err := ReadValue(data); err != nil {
			return err
		}
	}
	return nil
}

// Data returns the exact list bytes.
func (l List) Data() []byte {
	return l.data
}

// Element returns a list element data or nil.
func (l List) Element(i int) []byte {
	return l.element(i)
}

// Len returns the number of elements in the list.
func (l List) Len() int {
	return l.len()
}

// Getters

func (l List) Bool(i int) bool {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return false
	case end > int(l.body):
		return false
	}

	b := l.data[start:end]
	v, _ := readBool(b)
	return v
}

func (l List) Int8(i int) int8 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _ := readInt8(b)
	return v
}

func (l List) Int16(i int) int16 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _ := readInt16(b)
	return v
}

func (l List) Int32(i int) int32 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _ := readInt32(b)
	return v
}

func (l List) Int64(i int) int64 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _ := readInt64(b)
	return v
}

func (l List) Uint8(i int) uint8 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _ := readUint8(b)
	return v
}

func (l List) Uint16(i int) uint16 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _ := readUint16(b)
	return v
}

func (l List) Uint32(i int) uint32 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _ := readUint32(b)
	return v
}

func (l List) Uint64(i int) uint64 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _ := readUint64(b)
	return v
}

func (l List) Float32(i int) float32 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _ := readFloat32(b)
	return v
}

func (l List) Float64(i int) float64 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _ := readFloat64(b)
	return v
}

func (l List) Bytes(i int) []byte {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return nil
	case end > int(l.body):
		return nil
	}

	b := l.data[start:end]
	v, _ := readBytes(b)
	return v
}

func (l List) String(i int) string {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return ""
	case end > int(l.body):
		return ""
	}

	b := l.data[start:end]
	v, _ := readString(b)
	return v
}

func (l List) List(i int) List {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return List{}
	case end > int(l.body):
		return List{}
	}

	b := l.data[start:end]
	v, _ := GetList(b)
	return v
}

func (l List) Message(i int) MessageData {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return MessageData{}
	case end > int(l.body):
		return MessageData{}
	}

	b := l.data[start:end]
	v, _ := NewMessageData(b)
	return v
}
