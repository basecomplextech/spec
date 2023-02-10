package spec

import (
	"fmt"

	"github.com/complex1tech/baselibrary/types"
	"github.com/complex1tech/spec/encoding"
)

// Value is a raw value.
type Value []byte

// NewValue returns a new value from bytes or nil when not a value.
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

// ParseValue recursively parses and returns a value.
func ParseValue(b []byte) (_ Value, n int, err error) {
	t, n, err := encoding.DecodeType(b)
	if err != nil {
		return
	}

	switch t {
	case TypeTrue, TypeFalse:
		// pass

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

// Byte decodes and returns a byte or 0.
func (v Value) Byte() byte {
	p, _, _ := encoding.DecodeByte(v)
	return p
}

// Int

// Int16 decodes and returns int16 or 0.
func (v Value) Int16() int16 {
	p, _, _ := encoding.DecodeInt16(v)
	return p
}

// Int32 decodes and returns int32 or 0.
func (v Value) Int32() int32 {
	p, _, _ := encoding.DecodeInt32(v)
	return p
}

// Int64 decodes and returns int64 or 0.
func (v Value) Int64() int64 {
	p, _, _ := encoding.DecodeInt64(v)
	return p
}

// Uint

// Uint16 decodes and returns uint16 or 0.
func (v Value) Uint16() uint16 {
	p, _, _ := encoding.DecodeUint16(v)
	return p
}

// Uint32 decodes and returns uint32 or 0.
func (v Value) Uint32() uint32 {
	p, _, _ := encoding.DecodeUint32(v)
	return p
}

// Uint64 decodes and returns uint64 or 0.
func (v Value) Uint64() uint64 {
	p, _, _ := encoding.DecodeUint64(v)
	return p
}

// Float

// Float32 decodes and returns float32 or 0.
func (v Value) Float32() float32 {
	p, _, _ := encoding.DecodeFloat32(v)
	return p
}

// Float64 decodes and returns float64 or 0.
func (v Value) Float64() float64 {
	p, _, _ := encoding.DecodeFloat64(v)
	return p
}

// Bin

// Bin64 decodes and returns bin64 or a zero value.
func (v Value) Bin64() types.Bin64 {
	p, _, _ := encoding.DecodeBin64(v)
	return p
}

// Bin128 decodes and returns bin128 or a zero value.
func (v Value) Bin128() types.Bin128 {
	p, _, _ := encoding.DecodeBin128(v)
	return p
}

// Bin256 decodes and returns bin256 or a zero value.
func (v Value) Bin256() types.Bin256 {
	p, _, _ := encoding.DecodeBin256(v)
	return p
}

// Bytes/string

// Bytes decodes and returns bytes or nil.
func (v Value) Bytes() types.BytesView {
	p, _, _ := encoding.DecodeBytes(v)
	return p
}

// String decodes and returns string or an empty string.
func (v Value) String() types.StringView {
	p, _, _ := encoding.DecodeString(v)
	return p
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
