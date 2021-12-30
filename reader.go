package spec

// Reader reads a value from a byte slice.
type Reader struct {
	b []byte
}

// NewReader returns a new value reader.
func NewReader(b []byte) Reader {
	return Reader{b}
}

func (r Reader) ReadType() (Type, error) {
	return readType(r.b)
}

func (r Reader) ReadNil() (bool, error) {
	v, err := readType(r.b)
	if err != nil {
		return false, err
	}
	return v == TypeNil, nil
}

func (r Reader) ReadBool() (bool, error) {
	return readBool(r.b)
}

func (r Reader) Byte() (byte, error) {
	return readByte(r.b)
}

func (r Reader) Int8() (int8, error) {
	return readInt8(r.b)
}

func (r Reader) Int16() (int16, error) {
	return readInt16(r.b)
}

func (r Reader) Int32() (int32, error) {
	return readInt32(r.b)
}

func (r Reader) Int64() (int64, error) {
	return readInt64(r.b)
}

func (r Reader) Uint8() (uint8, error) {
	return readUint8(r.b)
}

func (r Reader) Uint16() (uint16, error) {
	return readUint16(r.b)
}

func (r Reader) Uint32() (uint32, error) {
	return readUint32(r.b)
}

func (r Reader) Uint64() (uint64, error) {
	return readUint64(r.b)
}

func (r Reader) Float32() (float32, error) {
	return readFloat32(r.b)
}

func (r Reader) Float64() (float64, error) {
	return readFloat64(r.b)
}

func (r Reader) Bytes() ([]byte, error) {
	return readBytes(r.b)
}

func (r Reader) String() (string, error) {
	return readString(r.b)
}

func (r Reader) List() (ListData, error) {
	return NewListData(r.b)
}

func (r Reader) Message() (MessageData, error) {
	return NewMessageData(r.b)
}
