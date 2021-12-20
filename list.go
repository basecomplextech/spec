package spec

type List struct {
	buffer []byte
	table  listTable
	data   []byte
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
	return l.buffer
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
	off := l.table.offset(i)
	switch {
	case off < 0:
		return nil
	case off > len(l.data):
		return nil
	}
	return l.data[:off]
}

// Len returns the number of elements in the list.
func (l List) Len() int {
	return l.table.count()
}

// Getters

func (l List) Bool(i int) bool {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return false
	case off > len(l.data):
		return false
	}

	b := l.data[:off]
	v, _ := ReadBool(b)
	return v
}

func (l List) Int8(i int) int8 {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return 0
	case off > len(l.data):
		return 0
	}

	b := l.data[:off]
	v, _ := ReadInt8(b)
	return v
}

func (l List) Int16(i int) int16 {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return 0
	case off > len(l.data):
		return 0
	}

	b := l.data[:off]
	v, _ := ReadInt16(b)
	return v
}

func (l List) Int32(i int) int32 {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return 0
	case off > len(l.data):
		return 0
	}

	b := l.data[:off]
	v, _ := ReadInt32(b)
	return v
}

func (l List) Int64(i int) int64 {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return 0
	case off > len(l.data):
		return 0
	}

	b := l.data[:off]
	v, _ := ReadInt64(b)
	return v
}

func (l List) UInt8(i int) uint8 {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return 0
	case off > len(l.data):
		return 0
	}

	b := l.data[:off]
	v, _ := ReadUInt8(b)
	return v
}

func (l List) UInt16(i int) uint16 {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return 0
	case off > len(l.data):
		return 0
	}

	b := l.data[:off]
	v, _ := ReadUInt16(b)
	return v
}

func (l List) UInt32(i int) uint32 {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return 0
	case off > len(l.data):
		return 0
	}

	b := l.data[:off]
	v, _ := ReadUInt32(b)
	return v
}

func (l List) UInt64(i int) uint64 {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return 0
	case off > len(l.data):
		return 0
	}

	b := l.data[:off]
	v, _ := ReadUInt64(b)
	return v
}

func (l List) Float32(i int) float32 {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return 0
	case off > len(l.data):
		return 0
	}

	b := l.data[:off]
	v, _ := ReadFloat32(b)
	return v
}

func (l List) Float64(i int) float64 {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return 0
	case off > len(l.data):
		return 0
	}

	b := l.data[:off]
	v, _ := ReadFloat64(b)
	return v
}

func (l List) Bytes(i int) []byte {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return nil
	case off > len(l.data):
		return nil
	}

	b := l.data[:off]
	v, _ := ReadBytes(b)
	return v
}

func (l List) String(i int) string {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return ""
	case off > len(l.data):
		return ""
	}

	b := l.data[:off]
	v, _ := ReadString(b)
	return v
}

func (l List) List(i int) List {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return List{}
	case off > len(l.data):
		return List{}
	}

	b := l.data[:off]
	v, _ := GetList(b)
	return v
}

func (l List) Message(i int) Message {
	off := l.table.offset(i)
	switch {
	case off <= 0:
		return Message{}
	case off > len(l.data):
		return Message{}
	}

	b := l.data[:off]
	v, _ := GetMessage(b)
	return v
}
