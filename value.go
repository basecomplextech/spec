package spec

type Value []byte

// GetValue parses and returns a value, but does not validate it.
func GetValue(b []byte) (Value, error) {
	t, err := readType(b)
	if err != nil {
		return Value{}, err
	}
	if err := checkType(t); err != nil {
		return Value{}, err
	}
	return Value(b), nil
}

// ReadValue reads, recursively validates and returns a value.
func ReadValue(b []byte) (Value, error) {
	v := Value(b)
	if err := v.Validate(); err != nil {
		return Value{}, err
	}
	return v, nil
}

func (v Value) Validate() error {
	t, err := readType(v)
	if err != nil {
		return err
	}

	switch t {
	case TypeNil, TypeTrue, TypeFalse:
		return nil

	case TypeInt8:
		_, err = readInt8(v)
	case TypeInt16:
		_, err = readInt16(v)
	case TypeInt32:
		_, err = readInt32(v)
	case TypeInt64:
		_, err = readInt64(v)

	case TypeUint8:
		_, err = readUint8(v)
	case TypeUint16:
		_, err = readUint16(v)
	case TypeUint32:
		_, err = readUint32(v)
	case TypeUint64:
		_, err = readUint64(v)

	case TypeFloat32:
		_, err = readFloat32(v)
	case TypeFloat64:
		_, err = readFloat64(v)

	case TypeBytes:
		_, err = readBytes(v)
	case TypeString:
		_, err = readString(v)

	case TypeList:
		_, err = ReadList(v)
	case TypeMessage:
		_, err = ReadMessage(v)
	}
	return err
}

func (v Value) Type() Type {
	p, _ := readType(v)
	return p
}

func (v Value) Nil() bool {
	p, _ := readBool(v)
	return p
}

func (v Value) Bool() bool {
	p, _ := readBool(v)
	return p
}

func (v Value) Byte() byte {
	p, _ := readByte(v)
	return p
}

func (v Value) Int8() int8 {
	p, _ := readInt8(v)
	return p
}

func (v Value) Int16() int16 {
	p, _ := readInt16(v)
	return p
}

func (v Value) Int32() int32 {
	p, _ := readInt32(v)
	return p
}

func (v Value) Int64() int64 {
	p, _ := readInt64(v)
	return p
}

func (v Value) Uint8() uint8 {
	p, _ := readUint8(v)
	return p
}

func (v Value) Uint16() uint16 {
	p, _ := readUint16(v)
	return p
}

func (v Value) Uint32() uint32 {
	p, _ := readUint32(v)
	return p
}

func (v Value) Uint64() uint64 {
	p, _ := readUint64(v)
	return p
}

func (v Value) Float32() float32 {
	p, _ := readFloat32(v)
	return p
}

func (v Value) Float64() float64 {
	p, _ := readFloat64(v)
	return p
}

func (v Value) Bytes() []byte {
	p, _ := readBytes(v)
	return p
}

func (v Value) String() string {
	p, _ := readString(v)
	return p
}

func (v Value) List() List {
	p, _ := GetList(v)
	return p
}

func (v Value) Message() Message {
	p, _ := GetMessage(v)
	return p
}
