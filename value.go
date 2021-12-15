package protocol

type Value struct {
	buf buffer
}

func ReadValue(b []byte) Value {
	buf := buffer(b)
	return Value{buf: buf}
}

func (v Value) Type() Type {
	return v.buf.peekType()
}

func (v Value) Nil() bool {
	return v.buf.peekType() == TypeNil
}

func (v Value) Bool() bool {
	t, _ := v.buf.type_()
	if t != TypeTrue {
		return false
	}
	return true
}

func (v Value) Byte() byte {
	return v.UInt8()
}

func (v Value) Int8() int8 {
	t, b := v.buf.type_()
	if t != TypeInt8 {
		return 0
	}

	x, _ := b.int8()
	return x
}

func (v Value) Int16() int16 {
	t, b := v.buf.type_()
	if t != TypeInt16 {
		return 0
	}

	x, _ := b.int16()
	return x
}

func (v Value) Int32() int32 {
	t, b := v.buf.type_()
	if t != TypeInt32 {
		return 0
	}

	x, _ := b.int32()
	return x
}

func (v Value) Int64() int64 {
	t, b := v.buf.type_()
	if t != TypeInt64 {
		return 0
	}

	x, _ := b.int64()
	return x
}

func (v Value) UInt8() uint8 {
	t, b := v.buf.type_()
	if t != TypeUInt8 {
		return 0
	}

	x, _ := b.uint8()
	return x
}

func (v Value) UInt16() uint16 {
	t, b := v.buf.type_()
	if t != TypeUInt16 {
		return 0
	}

	x, _ := b.uint16()
	return x
}

func (v Value) UInt32() uint32 {
	t, b := v.buf.type_()
	if t != TypeUInt32 {
		return 0
	}

	x, _ := b.uint32()
	return x
}

func (v Value) UInt64() uint64 {
	t, b := v.buf.type_()
	if t != TypeUInt64 {
		return 0
	}

	x, _ := b.uint64()
	return x
}

func (v Value) Float32() float32 {
	t, b := v.buf.type_()
	if t != TypeFloat32 {
		return 0
	}

	x, _ := b.float32()
	return x
}

func (v Value) Float64() float64 {
	t, b := v.buf.type_()
	if t != TypeFloat64 {
		return 0
	}

	x, _ := b.float64()
	return x
}

func (v Value) Bytes() []byte {
	t, b := v.buf.type_()
	if t != TypeBytes {
		return nil
	}

	size, b := b.size()
	x, _ := b.bytes(size)
	return x
}

func (v Value) String() string {
	t, b := v.buf.type_()
	if t != TypeString {
		return ""
	}

	size, b := b.size()
	x, _ := b.string(size)
	return x
}

func (v Value) List() List {
	return ReadList(v.buf)
}

func (v Value) Message() Message {
	return ReadMessage(v.buf)
}
