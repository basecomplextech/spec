package spec

import (
	"fmt"

	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
)

// Data is a raw value data.
type Data []byte

// NewData reads and returns value data, but does not validate its children.
func NewData(b []byte) (Data, error) {
	t, _, err := readType(b)
	if err != nil {
		return Data{}, err
	}
	if err := checkType(t); err != nil {
		return Data{}, err
	}
	return Data(b), nil
}

// ReadData reads and returns value data, and recursively validates its children.
func ReadData(b []byte) (Data, error) {
	d := Data(b)
	if err := d.validate(); err != nil {
		return Data{}, err
	}
	return d, nil
}

func (d Data) Type() Type {
	v, _, _ := readType(d)
	return v
}

func (d Data) Nil() bool {
	v, _, _ := readType(d)
	return v == TypeNil
}

func (d Data) Bool() bool {
	v, _, _ := readBool(d)
	return v
}

func (d Data) Byte() byte {
	v, _, _ := readByte(d)
	return v
}

func (d Data) Int8() int8 {
	v, _, _ := readInt8(d)
	return v
}

func (d Data) Int16() int16 {
	v, _, _ := readInt16(d)
	return v
}

func (d Data) Int32() int32 {
	v, _, _ := readInt32(d)
	return v
}

func (d Data) Int64() int64 {
	v, _, _ := readInt64(d)
	return v
}

func (d Data) Uint8() uint8 {
	v, _, _ := readUint8(d)
	return v
}

func (d Data) Uint16() uint16 {
	v, _, _ := readUint16(d)
	return v
}

func (d Data) Uint32() uint32 {
	v, _, _ := readUint32(d)
	return v
}

func (d Data) Uint64() uint64 {
	v, _, _ := readUint64(d)
	return v
}

func (d Data) U128() u128.U128 {
	v, _, _ := readU128(d)
	return v
}

func (d Data) U256() u256.U256 {
	v, _, _ := readU256(d)
	return v
}

func (d Data) Float32() float32 {
	v, _, _ := readFloat32(d)
	return v
}

func (d Data) Float64() float64 {
	v, _, _ := readFloat64(d)
	return v
}

func (d Data) Bytes() []byte {
	v, _, _ := readBytes(d)
	return v
}

func (d Data) String() string {
	v, _, _ := readString(d)
	return v
}

func (d Data) List() List {
	v, _ := NewList(d)
	return v
}

func (d Data) Message() Message {
	v, _ := NewMessage(d)
	return v
}

// private

func (d Data) validate() error {
	t, _, err := readType(d)
	if err != nil {
		return err
	}

	switch t {
	default:
		return fmt.Errorf("unsupported type %v", t)

	case TypeNil, TypeTrue, TypeFalse:
		return nil

	case TypeInt8:
		_, _, err = readInt8(d)
	case TypeInt16:
		_, _, err = readInt16(d)
	case TypeInt32:
		_, _, err = readInt32(d)
	case TypeInt64:
		_, _, err = readInt64(d)

	case TypeUint8:
		_, _, err = readUint8(d)
	case TypeUint16:
		_, _, err = readUint16(d)
	case TypeUint32:
		_, _, err = readUint32(d)
	case TypeUint64:
		_, _, err = readUint64(d)

	case TypeU128:
		_, _, err = readU128(d)
	case TypeU256:
		_, _, err = readU256(d)

	case TypeFloat32:
		_, _, err = readFloat32(d)
	case TypeFloat64:
		_, _, err = readFloat64(d)

	case TypeBytes:
		_, _, err = readBytes(d)
	case TypeString:
		_, _, err = readString(d)

	case TypeList:
		_, err = ReadList(d)
	case TypeMessage,
		TypeMessageBig:
		_, err = ReadMessage(d)

		// TODO: Uncomment
		// case TypeStruct:
		// 	_, err = ReadStruct(d)
	}
	return err
}
