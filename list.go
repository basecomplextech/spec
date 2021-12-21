package spec

type List struct {
	data  []byte
	body  []byte
	table listTable
	big   bool
}

// GetList parses and returns a list, but does not validate it.
func GetList(b []byte) (List, error) {
	return readList(b)
}

// ReadList reads, recursively validates and returns a list.
func ReadList(b []byte) (List, error) {
	l, err := readList(b)
	if err != nil {
		return List{}, err
	}
	if err := l.Validate(); err != nil {
		return List{}, err
	}
	return l, nil
}

// Data returns the exact list bytes.
func (l List) Data() []byte {
	return l.data
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

// Element returns a list element data or nil.
func (l List) Element(i int) []byte {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return nil
	case end > len(l.body):
		return nil
	}
	return l.body[start:end]
}

// Len returns the number of elements in the list.
func (l List) Len() int {
	return l.table.count(l.big)
}

// Getters

func (l List) Bool(i int) bool {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return false
	case end > len(l.body):
		return false
	}

	b := l.body[start:end]
	v, _ := ReadBool(b)
	return v
}

func (l List) Int8(i int) int8 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > len(l.body):
		return 0
	}

	b := l.body[start:end]
	v, _ := ReadInt8(b)
	return v
}

func (l List) Int16(i int) int16 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > len(l.body):
		return 0
	}

	b := l.body[start:end]
	v, _ := ReadInt16(b)
	return v
}

func (l List) Int32(i int) int32 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > len(l.body):
		return 0
	}

	b := l.body[start:end]
	v, _ := ReadInt32(b)
	return v
}

func (l List) Int64(i int) int64 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > len(l.body):
		return 0
	}

	b := l.body[start:end]
	v, _ := ReadInt64(b)
	return v
}

func (l List) UInt8(i int) uint8 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > len(l.body):
		return 0
	}

	b := l.body[start:end]
	v, _ := ReadUInt8(b)
	return v
}

func (l List) UInt16(i int) uint16 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > len(l.body):
		return 0
	}

	b := l.body[start:end]
	v, _ := ReadUInt16(b)
	return v
}

func (l List) UInt32(i int) uint32 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > len(l.body):
		return 0
	}

	b := l.body[start:end]
	v, _ := ReadUInt32(b)
	return v
}

func (l List) UInt64(i int) uint64 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > len(l.body):
		return 0
	}

	b := l.body[start:end]
	v, _ := ReadUInt64(b)
	return v
}

func (l List) Float32(i int) float32 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > len(l.body):
		return 0
	}

	b := l.body[start:end]
	v, _ := ReadFloat32(b)
	return v
}

func (l List) Float64(i int) float64 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > len(l.body):
		return 0
	}

	b := l.body[start:end]
	v, _ := ReadFloat64(b)
	return v
}

func (l List) Bytes(i int) []byte {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return nil
	case end > len(l.body):
		return nil
	}

	b := l.body[start:end]
	v, _ := readBytes(b)
	return v
}

func (l List) String(i int) string {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return ""
	case end > len(l.body):
		return ""
	}

	b := l.body[start:end]
	v, _ := readString(b)
	return v
}

func (l List) List(i int) List {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return List{}
	case end > len(l.body):
		return List{}
	}

	b := l.body[start:end]
	v, _ := GetList(b)
	return v
}

func (l List) Message(i int) Message {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return Message{}
	case end > len(l.body):
		return Message{}
	}

	b := l.body[start:end]
	v, _ := GetMessage(b)
	return v
}
