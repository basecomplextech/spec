package spec

type List struct {
	buffer []byte
	table  listTable
	data   []byte
}

// Data returns the exact list bytes.
func (l List) Data() []byte {
	return l.buffer
}

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
	case off < 0:
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
	case off < 0:
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
	case off < 0:
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
	case off < 0:
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
	case off < 0:
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
	case off < 0:
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
	case off < 0:
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
	case off < 0:
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
	case off < 0:
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
	case off < 0:
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
	case off < 0:
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
	case off < 0:
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
	case off < 0:
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
	case off < 0:
		return List{}
	case off > len(l.data):
		return List{}
	}

	b := l.data[:off]
	v, _ := ReadList(b)
	return v
}

func (l List) Message(i int) Message {
	off := l.table.offset(i)
	switch {
	case off < 0:
		return Message{}
	case off > len(l.data):
		return Message{}
	}

	b := l.data[:off]
	v, _ := ReadMessage(b)
	return v
}
