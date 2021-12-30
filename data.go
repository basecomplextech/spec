package spec

// Data is a raw value data.
type Data []byte

// NewData parses and returns value data, but does not validate it.
func NewData(b []byte) (Data, error) {
	t, err := readType(b)
	if err != nil {
		return Data{}, err
	}
	if err := checkType(t); err != nil {
		return Data{}, err
	}
	return Data(b), nil
}

// ReadData reads, recursively validates and returns value data.
func ReadData(b []byte) (Data, error) {
	d := Data(b)
	if err := d.validate(); err != nil {
		return Data{}, err
	}
	return d, nil
}

func (d Data) validate() error {
	t, err := readType(d)
	if err != nil {
		return err
	}

	switch t {
	case TypeNil, TypeTrue, TypeFalse:
		return nil

	case TypeInt8:
		_, err = readInt8(d)
	case TypeInt16:
		_, err = readInt16(d)
	case TypeInt32:
		_, err = readInt32(d)
	case TypeInt64:
		_, err = readInt64(d)

	case TypeUint8:
		_, err = readUint8(d)
	case TypeUint16:
		_, err = readUint16(d)
	case TypeUint32:
		_, err = readUint32(d)
	case TypeUint64:
		_, err = readUint64(d)

	case TypeFloat32:
		_, err = readFloat32(d)
	case TypeFloat64:
		_, err = readFloat64(d)

	case TypeBytes:
		_, err = readBytes(d)
	case TypeString:
		_, err = readString(d)

	case TypeList:
		_, err = ReadListData(d)
	case TypeMessage:
		_, err = ReadMessageData(d)
	}
	return err
}

func (d Data) Type() Type {
	v, _ := readType(d)
	return v
}

func (d Data) Nil() bool {
	v, _ := readType(d)
	return v == TypeNil
}

func (d Data) Bool() bool {
	v, _ := readBool(d)
	return v
}

func (d Data) Byte() byte {
	v, _ := readByte(d)
	return v
}

func (d Data) Int8() int8 {
	v, _ := readInt8(d)
	return v
}

func (d Data) Int16() int16 {
	v, _ := readInt16(d)
	return v
}

func (d Data) Int32() int32 {
	v, _ := readInt32(d)
	return v
}

func (d Data) Int64() int64 {
	v, _ := readInt64(d)
	return v
}

func (d Data) Uint8() uint8 {
	v, _ := readUint8(d)
	return v
}

func (d Data) Uint16() uint16 {
	v, _ := readUint16(d)
	return v
}

func (d Data) Uint32() uint32 {
	v, _ := readUint32(d)
	return v
}

func (d Data) Uint64() uint64 {
	v, _ := readUint64(d)
	return v
}

func (d Data) Float32() float32 {
	v, _ := readFloat32(d)
	return v
}

func (d Data) Float64() float64 {
	v, _ := readFloat64(d)
	return v
}

func (d Data) Bytes() []byte {
	v, _ := readBytes(d)
	return v
}

func (d Data) String() string {
	v, _ := readString(d)
	return v
}

func (d Data) List() ListData {
	v, _ := NewListData(d)
	return v
}

func (d Data) Message() MessageData {
	v, _ := NewMessageData(d)
	return v
}
