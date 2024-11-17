// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package types

import (
	"fmt"

	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/spec/internal/decode"
	"github.com/basecomplextech/spec/internal/format"
)

// Value is a raw value.
type Value []byte

// NewValue returns a new value from bytes or nil when not valid.
func NewValue(b []byte) Value {
	_, n, err := decode.DecodeTypeSize(b)
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
	_, n, err := decode.DecodeTypeSize(b)
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
	typ, n, err := decode.DecodeType(b)
	if err != nil {
		return
	}

	switch typ {
	case format.TypeTrue, format.TypeFalse:
		// Pass

	case format.TypeByte:
		_, n, err = decode.DecodeByte(b)

	case format.TypeInt16:
		_, n, err = decode.DecodeInt16(b)
	case format.TypeInt32:
		_, n, err = decode.DecodeInt32(b)
	case format.TypeInt64:
		_, n, err = decode.DecodeInt64(b)

	case format.TypeUint16:
		_, n, err = decode.DecodeUint16(b)
	case format.TypeUint32:
		_, n, err = decode.DecodeUint32(b)
	case format.TypeUint64:
		_, n, err = decode.DecodeUint64(b)

	case format.TypeBin64:
		_, n, err = decode.DecodeBin64(b)
	case format.TypeBin128:
		_, n, err = decode.DecodeBin128(b)
	case format.TypeBin256:
		_, n, err = decode.DecodeBin256(b)

	case format.TypeFloat32:
		_, n, err = decode.DecodeFloat32(b)
	case format.TypeFloat64:
		_, n, err = decode.DecodeFloat64(b)

	case format.TypeBytes:
		_, n, err = decode.DecodeBytes(b)
	case format.TypeString:
		_, n, err = decode.DecodeString(b)

	case format.TypeList, format.TypeBigList:
		_, n, err = ParseList(b)

	case format.TypeMessage, format.TypeBigMessage:
		_, n, err = ParseMessage(b)

	case format.TypeStruct:
		_, n, err = decode.DecodeStruct(b)

	default:
		n, err = 0, fmt.Errorf("unsupported type %d", typ)
	}
	if err != nil {
		return nil, n, err
	}

	return b[len(b)-n:], n, nil
}

// Types

// Type decodes and returns a type or undefined.
func (v Value) Type() format.Type {
	p, _, _ := decode.DecodeType(v)
	return p
}

// Bool decodes and returns a bool or false.
func (v Value) Bool() bool {
	p, _, _ := decode.DecodeBool(v)
	return p
}

// BoolErr decodes and returns a bool or an error.
func (v Value) BoolErr() (bool, error) {
	p, _, err := decode.DecodeBool(v)
	return p, err
}

// Byte decodes and returns a byte or 0.
func (v Value) Byte() byte {
	p, _, _ := decode.DecodeByte(v)
	return p
}

// ByteErr decodes and returns a byte or an error.
func (v Value) ByteErr() (byte, error) {
	p, _, err := decode.DecodeByte(v)
	return p, err
}

// Int

// Int16 decodes and returns an int16 or 0.
func (v Value) Int16() int16 {
	p, _, _ := decode.DecodeInt16(v)
	return p
}

// Int16Err decodes and returns an int16 or an error.
func (v Value) Int16Err() (int16, error) {
	p, _, err := decode.DecodeInt16(v)
	return p, err
}

// Int32 decodes and returns an int32 or 0.
func (v Value) Int32() int32 {
	p, _, _ := decode.DecodeInt32(v)
	return p
}

// Int32Err decodes and returns an int32 or an error.
func (v Value) Int32Err() (int32, error) {
	p, _, err := decode.DecodeInt32(v)
	return p, err
}

// Int64 decodes and returns an int64 or 0.
func (v Value) Int64() int64 {
	p, _, _ := decode.DecodeInt64(v)
	return p
}

// Int64Err decodes and returns an int64 or an error.
func (v Value) Int64Err() (int64, error) {
	p, _, err := decode.DecodeInt64(v)
	return p, err
}

// Uint

// Uint16 decodes and returns a uint16 or 0.
func (v Value) Uint16() uint16 {
	p, _, _ := decode.DecodeUint16(v)
	return p
}

// Uint16Err decodes and returns a uint16 or an error.
func (v Value) Uint16Err() (uint16, error) {
	p, _, err := decode.DecodeUint16(v)
	return p, err
}

// Uint32 decodes and returns a uint32 or 0.
func (v Value) Uint32() uint32 {
	p, _, _ := decode.DecodeUint32(v)
	return p
}

// Uint32Err decodes and returns a uint32 or an error.
func (v Value) Uint32Err() (uint32, error) {
	p, _, err := decode.DecodeUint32(v)
	return p, err
}

// Uint64 decodes and returns a uint64 or 0.
func (v Value) Uint64() uint64 {
	p, _, _ := decode.DecodeUint64(v)
	return p
}

// Uint64Err decodes and returns a uint64 or an error.
func (v Value) Uint64Err() (uint64, error) {
	p, _, err := decode.DecodeUint64(v)
	return p, err
}

// Float

// Float32 decodes and returns a float32 or 0.
func (v Value) Float32() float32 {
	p, _, _ := decode.DecodeFloat32(v)
	return p
}

// Float32Err decodes and returns a float32 or an error.
func (v Value) Float32Err() (float32, error) {
	p, _, err := decode.DecodeFloat32(v)
	return p, err
}

// Float64 decodes and returns a float64 or 0.
func (v Value) Float64() float64 {
	p, _, _ := decode.DecodeFloat64(v)
	return p
}

// Float64Err decodes and returns a float64 or an error.
func (v Value) Float64Err() (float64, error) {
	p, _, err := decode.DecodeFloat64(v)
	return p, err
}

// Bin

// Bin64 decodes and returns a bin64 or a zero value.
func (v Value) Bin64() bin.Bin64 {
	p, _, _ := decode.DecodeBin64(v)
	return p
}

// Bin64Err decodes and returns bin96 or an error.
func (v Value) Bin64Err() (bin.Bin64, error) {
	p, _, err := decode.DecodeBin64(v)
	return p, err
}

// Bin128 decodes and returns a bin128 or a zero value.
func (v Value) Bin128() bin.Bin128 {
	p, _, _ := decode.DecodeBin128(v)
	return p
}

// Bin128Err decodes and returns a bin128 or an error.
func (v Value) Bin128Err() (bin.Bin128, error) {
	p, _, err := decode.DecodeBin128(v)
	return p, err
}

// Bin256 decodes and returns a bin256 or a zero value.
func (v Value) Bin256() bin.Bin256 {
	p, _, _ := decode.DecodeBin256(v)
	return p
}

// Bin256Err decodes and returns a bin256 or an error.
func (v Value) Bin256Err() (bin.Bin256, error) {
	p, _, err := decode.DecodeBin256(v)
	return p, err
}

// Bytes/string

// Bytes decodes and returns bytes or nil.
func (v Value) Bytes() format.Bytes {
	p, _, _ := decode.DecodeBytes(v)
	return p
}

// BytesErr decodes and returns bytes or an error.
func (v Value) BytesErr() (format.Bytes, error) {
	p, _, err := decode.DecodeBytes(v)
	return p, err
}

// String decodes and returns a string or an empty string.
func (v Value) String() format.String {
	p, _, _ := decode.DecodeString(v)
	return format.String(p)
}

// StringErr decodes and returns a string or an error.
func (v Value) StringErr() (format.String, error) {
	p, _, err := decode.DecodeString(v)
	return format.String(p), err
}

// List/message

// List decodes and returns a list or an empty list.
func (v Value) List() List {
	return OpenList(v)
}

// ListErr decodes and returns a list or an error.
func (v Value) ListErr() (List, error) {
	return OpenListErr(v)
}

// Message decodes and returns a message or an empty message.
func (v Value) Message() Message {
	return OpenMessage(v)
}

// MessageErr decodes and returns a message or an error.
func (v Value) MessageErr() (Message, error) {
	return OpenMessageErr(v)
}
