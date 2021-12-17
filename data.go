package spec

type Data []byte

func ReadData(b []byte) Data { return Data(b) }

func (d Data) Type() Type {
	return ReadType(d)
}

func (d Data) Nil() bool {
	return ReadType(d) == TypeNil
}

func (d Data) Bool() bool {
	return ReadBool(d)
}

func (d Data) Byte() byte {
	return ReadByte(d)
}

func (d Data) Int8() int8 {
	return ReadInt8(d)
}

func (d Data) Int16() int16 {
	return ReadInt16(d)
}

func (d Data) Int32() int32 {
	return ReadInt32(d)
}

func (d Data) Int64() int64 {
	return ReadInt64(d)
}

func (d Data) UInt8() uint8 {
	return ReadUInt8(d)
}

func (d Data) UInt16() uint16 {
	return ReadUInt16(d)
}

func (d Data) UInt32() uint32 {
	return ReadUInt32(d)
}

func (d Data) UInt64() uint64 {
	return ReadUInt64(d)
}

func (d Data) Float32() float32 {
	return ReadFloat32(d)
}

func (d Data) Float64() float64 {
	return ReadFloat64(d)
}

func (d Data) Bytes() []byte {
	return ReadBytes(d)
}

func (d Data) String() string {
	return ReadString(d)
}

func (d Data) List() List {
	return ReadList(d)
}

func (d Data) Message() Message {
	return ReadMessage(d)
}
