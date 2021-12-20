package spec

type Value []byte

func ReadValue(b []byte) Value {
	return Value(b)
}

func (v Value) Type() Type {
	p, _ := ReadType(v)
	return p
}

func (v Value) Nil() bool {
	p, _ := ReadBool(v)
	return p
}

func (v Value) Bool() bool {
	p, _ := ReadBool(v)
	return p
}

func (v Value) Byte() byte {
	p, _ := ReadByte(v)
	return p
}

func (v Value) Int8() int8 {
	p, _ := ReadInt8(v)
	return p
}

func (v Value) Int16() int16 {
	p, _ := ReadInt16(v)
	return p
}

func (v Value) Int32() int32 {
	p, _ := ReadInt32(v)
	return p
}

func (v Value) Int64() int64 {
	p, _ := ReadInt64(v)
	return p
}

func (v Value) UInt8() uint8 {
	p, _ := ReadUInt8(v)
	return p
}

func (v Value) UInt16() uint16 {
	p, _ := ReadUInt16(v)
	return p
}

func (v Value) UInt32() uint32 {
	p, _ := ReadUInt32(v)
	return p
}

func (v Value) UInt64() uint64 {
	p, _ := ReadUInt64(v)
	return p
}

func (v Value) Float32() float32 {
	p, _ := ReadFloat32(v)
	return p
}

func (v Value) Float64() float64 {
	p, _ := ReadFloat64(v)
	return p
}

func (v Value) Bytes() []byte {
	p, _ := ReadBytes(v)
	return p
}

func (v Value) String() string {
	p, _ := ReadString(v)
	return p
}

func (v Value) List() List {
	p, _ := ReadList(v)
	return p
}

func (v Value) Message() Message {
	p, _ := ReadMessage(v)
	return p
}
