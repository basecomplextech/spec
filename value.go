package spec

import (
	"fmt"

	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
)

// Value is a raw value.
type Value []byte

// NewValue reads and returns a value or zero on an error.
func NewValue(b []byte) Value {
	t, _, err := ReadType(b)
	if err != nil {
		return Value{}
	}
	if err := checkType(t); err != nil {
		return Value{}
	}
	return Value(b)
}

// ReadValue reads, recursively validates and returns a value.
func ReadValue(b []byte) (Value, int, error) {
	t, n, err := ReadType(b)
	if err != nil {
		return Value{}, n, err
	}
	if err := checkType(t); err != nil {
		return Value{}, n, err
	}

	v := Value(b)
	if err := v.validate(); err != nil {
		return Value{}, n, err
	}
	return v, n, nil
}

func (v Value) Type() Type {
	p, _, _ := ReadType(v)
	return p
}

func (v Value) Nil() bool {
	p, _, _ := ReadType(v)
	return p == TypeNil
}

func (v Value) Bool() bool {
	p, _, _ := ReadBool(v)
	return p
}

func (v Value) Byte() byte {
	p, _, _ := ReadByte(v)
	return p
}

func (v Value) Int32() int32 {
	p, _, _ := ReadInt32(v)
	return p
}

func (v Value) Int64() int64 {
	p, _, _ := ReadInt64(v)
	return p
}

func (v Value) Uint32() uint32 {
	p, _, _ := ReadUint32(v)
	return p
}

func (v Value) Uint64() uint64 {
	p, _, _ := ReadUint64(v)
	return p
}

func (v Value) U128() u128.U128 {
	p, _, _ := ReadU128(v)
	return p
}

func (v Value) U256() u256.U256 {
	p, _, _ := ReadU256(v)
	return p
}

func (v Value) Float32() float32 {
	p, _, _ := ReadFloat32(v)
	return p
}

func (v Value) Float64() float64 {
	p, _, _ := ReadFloat64(v)
	return p
}

func (v Value) Bytes() []byte {
	p, _, _ := ReadBytes(v)
	return p
}

func (v Value) String() string {
	p, _, _ := ReadString(v)
	return p
}

func (v Value) Message() Message {
	return NewMessage(v)
}

// private

func (v Value) validate() error {
	t, _, err := ReadType(v)
	if err != nil {
		return err
	}

	switch t {
	default:
		return fmt.Errorf("unsupported type %d", t)

	case TypeNil, TypeTrue, TypeFalse:
		return nil
	case TypeByte:
		_, _, err = ReadByte(v)

	case TypeInt32:
		_, _, err = ReadInt32(v)
	case TypeInt64:
		_, _, err = ReadInt64(v)

	case TypeUint32:
		_, _, err = ReadUint32(v)
	case TypeUint64:
		_, _, err = ReadUint64(v)

	case TypeU128:
		_, _, err = ReadU128(v)
	case TypeU256:
		_, _, err = ReadU256(v)

	case TypeFloat32:
		_, _, err = ReadFloat32(v)
	case TypeFloat64:
		_, _, err = ReadFloat64(v)

	case TypeBytes:
		_, _, err = ReadBytes(v)
	case TypeString:
		_, _, err = ReadString(v)

	case TypeList, TypeListBig:
		_, _, err = ReadList(v, ReadValue)

	case TypeMessage, TypeMessageBig:
		_, _, err = ReadMessage(v)

		// TODO: Uncomment
		// case TypeStruct:
		// 	_, err = ReadStruct(v)
	}
	return err
}
