package spec

import (
	"fmt"

	"github.com/complex1tech/baselibrary/bin"
	"github.com/complex1tech/baselibrary/mod"
	"github.com/complex1tech/spec/encoding"
)

// Value is a raw value.
type Value []byte

// NewValue returns a new value from bytes or nil when not valid.
func NewValue(b []byte) Value {
	_, n, err := encoding.DecodeTypeSize(b)
	switch {
	case err != nil:
		return nil
	case len(b) < n:
		return nil
	}

	return b[len(b)-n:]
}

// NewValueErr returns a new value from bytes or an error when not valid.
func NewValueErr(b []byte) (Value, error) {
	_, n, err := encoding.DecodeTypeSize(b)
	switch {
	case err != nil:
		return Value{}, err
	case len(b) < n:
		return Value{}, err
	}

	return b[len(b)-n:], nil
}

// ParseValue recursively parses and returns a value.
func ParseValue(b []byte) (_ Value, n int, err error) {
	t, n, err := encoding.DecodeType(b)
	if err != nil {
		return
	}

	switch t {
	case TypeTrue, TypeFalse:
		// Pass

	case TypeByte:
		_, n, err = encoding.DecodeByte(b)

	case TypeInt16:
		_, n, err = encoding.DecodeInt16(b)
	case TypeInt32:
		_, n, err = encoding.DecodeInt32(b)
	case TypeInt64:
		_, n, err = encoding.DecodeInt64(b)

	case TypeUint16:
		_, n, err = encoding.DecodeUint16(b)
	case TypeUint32:
		_, n, err = encoding.DecodeUint32(b)
	case TypeUint64:
		_, n, err = encoding.DecodeUint64(b)

	case TypeBin64:
		_, n, err = encoding.DecodeBin64(b)
	case TypeBin128:
		_, n, err = encoding.DecodeBin128(b)
	case TypeBin256:
		_, n, err = encoding.DecodeBin256(b)

	case TypeFloat32:
		_, n, err = encoding.DecodeFloat32(b)
	case TypeFloat64:
		_, n, err = encoding.DecodeFloat64(b)

	case TypeBytes:
		_, n, err = encoding.DecodeBytes(b)
	case TypeString:
		_, n, err = encoding.DecodeString(b)

	case TypeList, TypeBigList:
		_, n, err = ParseList(b)

	case TypeMessage, TypeBigMessage:
		_, n, err = ParseMessage(b)

	case TypeStruct:
		_, n, err = encoding.DecodeStruct(b)

	default:
		n, err = 0, fmt.Errorf("unsupported type %d", t)
	}
	if err != nil {
		return nil, n, err
	}

	return b[len(b)-n:], n, nil
}

// Types

// Type decodes and returns a type or undefined.
func (v Value) Type() Type {
	p, _, _ := encoding.DecodeType(v)
	return p
}

// Bool decodes and returns a bool or false.
func (v Value) Bool() bool {
	p, _, _ := encoding.DecodeBool(v)
	return p
}

// BoolErr decodes and returns a bool or an error.
func (v Value) BoolErr() (bool, error) {
	p, _, err := encoding.DecodeBool(v)
	return p, err
}

// Byte decodes and returns a byte or 0.
func (v Value) Byte() byte {
	p, _, _ := encoding.DecodeByte(v)
	return p
}

// ByteErr decodes and returns a byte or an error.
func (v Value) ByteErr() (byte, error) {
	p, _, err := encoding.DecodeByte(v)
	return p, err
}

// Int

// Int16 decodes and returns an int16 or 0.
func (v Value) Int16() int16 {
	p, _, _ := encoding.DecodeInt16(v)
	return p
}

// Int16Err decodes and returns an int16 or an error.
func (v Value) Int16Err() (int16, error) {
	p, _, err := encoding.DecodeInt16(v)
	return p, err
}

// Int32 decodes and returns an int32 or 0.
func (v Value) Int32() int32 {
	p, _, _ := encoding.DecodeInt32(v)
	return p
}

// Int32Err decodes and returns an int32 or an error.
func (v Value) Int32Err() (int32, error) {
	p, _, err := encoding.DecodeInt32(v)
	return p, err
}

// Int64 decodes and returns an int64 or 0.
func (v Value) Int64() int64 {
	p, _, _ := encoding.DecodeInt64(v)
	return p
}

// Int64Err decodes and returns an int64 or an error.
func (v Value) Int64Err() (int64, error) {
	p, _, err := encoding.DecodeInt64(v)
	return p, err
}

// Uint

// Uint16 decodes and returns a uint16 or 0.
func (v Value) Uint16() uint16 {
	p, _, _ := encoding.DecodeUint16(v)
	return p
}

// Uint16Err decodes and returns a uint16 or an error.
func (v Value) Uint16Err() (uint16, error) {
	p, _, err := encoding.DecodeUint16(v)
	return p, err
}

// Uint32 decodes and returns a uint32 or 0.
func (v Value) Uint32() uint32 {
	p, _, _ := encoding.DecodeUint32(v)
	return p
}

// Uint32Err decodes and returns a uint32 or an error.
func (v Value) Uint32Err() (uint32, error) {
	p, _, err := encoding.DecodeUint32(v)
	return p, err
}

// Uint64 decodes and returns a uint64 or 0.
func (v Value) Uint64() uint64 {
	p, _, _ := encoding.DecodeUint64(v)
	return p
}

// Uint64Err decodes and returns a uint64 or an error.
func (v Value) Uint64Err() (uint64, error) {
	p, _, err := encoding.DecodeUint64(v)
	return p, err
}

// Float

// Float32 decodes and returns a float32 or 0.
func (v Value) Float32() float32 {
	p, _, _ := encoding.DecodeFloat32(v)
	return p
}

// Float32Err decodes and returns a float32 or an error.
func (v Value) Float32Err() (float32, error) {
	p, _, err := encoding.DecodeFloat32(v)
	return p, err
}

// Float64 decodes and returns a float64 or 0.
func (v Value) Float64() float64 {
	p, _, _ := encoding.DecodeFloat64(v)
	return p
}

// Float64Err decodes and returns a float64 or an error.
func (v Value) Float64Err() (float64, error) {
	p, _, err := encoding.DecodeFloat64(v)
	return p, err
}

// Bin

// Bin64 decodes and returns a bin64 or a zero value.
func (v Value) Bin64() bin.Bin64 {
	p, _, _ := encoding.DecodeBin64(v)
	return p
}

// Bin64Err decodes and returns bin96 or an error.
func (v Value) Bin64Err() (bin.Bin64, error) {
	p, _, err := encoding.DecodeBin64(v)
	return p, err
}

// Bin128 decodes and returns a bin128 or a zero value.
func (v Value) Bin128() bin.Bin128 {
	p, _, _ := encoding.DecodeBin128(v)
	return p
}

// Bin128Err decodes and returns a bin128 or an error.
func (v Value) Bin128Err() (bin.Bin128, error) {
	p, _, err := encoding.DecodeBin128(v)
	return p, err
}

// Bin256 decodes and returns a bin256 or a zero value.
func (v Value) Bin256() bin.Bin256 {
	p, _, _ := encoding.DecodeBin256(v)
	return p
}

// Bin256Err decodes and returns a bin256 or an error.
func (v Value) Bin256Err() (bin.Bin256, error) {
	p, _, err := encoding.DecodeBin256(v)
	return p, err
}

// Bytes/string

// Bytes decodes and returns bytes or nil.
func (v Value) Bytes() mod.Ext[[]byte] {
	p, _, _ := encoding.DecodeBytes(v)
	return mod.NewExt(p)
}

// BytesErr decodes and returns bytes or an error.
func (v Value) BytesErr() (mod.Ext[[]byte], error) {
	p, _, err := encoding.DecodeBytes(v)
	return mod.NewExt(p), err
}

// String decodes and returns a string or an empty string.
func (v Value) String() mod.Ext[string] {
	p, _, _ := encoding.DecodeString(v)
	return mod.NewExt(p)
}

// StringErr decodes and returns a string or an error.
func (v Value) StringErr() (mod.Ext[string], error) {
	p, _, err := encoding.DecodeString(v)
	return mod.NewExt(p), err
}

// List/message

// List decodes and returns a list or an empty list.
func (v Value) List() List {
	return NewList(v)
}

// Message decodes and returns a message or an empty message.
func (v Value) Message() Message {
	return NewMessage(v)
}
