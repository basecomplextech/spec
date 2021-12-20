package spec

type Data []byte

func ReadData(b []byte) Data { return Data(b) }

func (d Data) Type() Type {
	v, _ := ReadType(d)
	return v
}

func (d Data) Nil() bool {
	v, _ := ReadBool(d)
	return v
}

func (d Data) Bool() bool {
	v, _ := ReadBool(d)
	return v
}

func (d Data) Byte() byte {
	v, _ := ReadByte(d)
	return v
}

func (d Data) Int8() int8 {
	v, _ := ReadInt8(d)
	return v
}

func (d Data) Int16() int16 {
	v, _ := ReadInt16(d)
	return v
}

func (d Data) Int32() int32 {
	v, _ := ReadInt32(d)
	return v
}

func (d Data) Int64() int64 {
	v, _ := ReadInt64(d)
	return v
}

func (d Data) UInt8() uint8 {
	v, _ := ReadUInt8(d)
	return v
}

func (d Data) UInt16() uint16 {
	v, _ := ReadUInt16(d)
	return v
}

func (d Data) UInt32() uint32 {
	v, _ := ReadUInt32(d)
	return v
}

func (d Data) UInt64() uint64 {
	v, _ := ReadUInt64(d)
	return v
}

func (d Data) Float32() float32 {
	v, _ := ReadFloat32(d)
	return v
}

func (d Data) Float64() float64 {
	v, _ := ReadFloat64(d)
	return v
}

func (d Data) Bytes() []byte {
	v, _ := ReadBytes(d)
	return v
}

func (d Data) String() string {
	v, _ := ReadString(d)
	return v
}

func (d Data) List() List {
	v, _ := ReadList(d)
	return v
}

func (d Data) Message() Message {
	v, _ := ReadMessage(d)
	return v
}
