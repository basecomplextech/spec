// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package types

import (
	"fmt"

	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/spec/encoding"
	"github.com/basecomplextech/spec/internal/core"
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
	typ, n, err := encoding.DecodeType(b)
	if err != nil {
		return
	}

	switch typ {
	case core.TypeTrue, core.TypeFalse:
		// Pass

	case core.TypeByte:
		_, n, err = encoding.DecodeByte(b)

	case core.TypeInt16:
		_, n, err = encoding.DecodeInt16(b)
	case core.TypeInt32:
		_, n, err = encoding.DecodeInt32(b)
	case core.TypeInt64:
		_, n, err = encoding.DecodeInt64(b)

	case core.TypeUint16:
		_, n, err = encoding.DecodeUint16(b)
	case core.TypeUint32:
		_, n, err = encoding.DecodeUint32(b)
	case core.TypeUint64:
		_, n, err = encoding.DecodeUint64(b)

	case core.TypeBin64:
		_, n, err = encoding.DecodeBin64(b)
	case core.TypeBin128:
		_, n, err = encoding.DecodeBin128(b)
	case core.TypeBin256:
		_, n, err = encoding.DecodeBin256(b)

	case core.TypeFloat32:
		_, n, err = encoding.DecodeFloat32(b)
	case core.TypeFloat64:
		_, n, err = encoding.DecodeFloat64(b)

	case core.TypeBytes:
		_, n, err = encoding.DecodeBytes(b)
	case core.TypeString:
		_, n, err = encoding.DecodeString(b)

	case core.TypeList, core.TypeBigList:
		_, n, err = ParseList(b)

	case core.TypeMessage, core.TypeBigMessage:
		_, n, err = ParseMessage(b)

	case core.TypeStruct:
		_, n, err = encoding.DecodeStruct(b)

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
func (v Value) Type() core.Type {
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
func (v Value) Bytes() core.Bytes {
	p, _, _ := encoding.DecodeBytes(v)
	return p
}

// BytesErr decodes and returns bytes or an error.
func (v Value) BytesErr() (core.Bytes, error) {
	p, _, err := encoding.DecodeBytes(v)
	return p, err
}

// String decodes and returns a string or an empty string.
func (v Value) String() core.String {
	p, _, _ := encoding.DecodeString(v)
	return core.String(p)
}

// StringErr decodes and returns a string or an error.
func (v Value) StringErr() (core.String, error) {
	p, _, err := encoding.DecodeString(v)
	return core.String(p), err
}

// List/message

// List decodes and returns a list or an empty list.
func (v Value) List() List {
	return NewList(v)
}

// ListErr decodes and returns a list or an error.
func (v Value) ListErr() (List, error) {
	return NewListErr(v)
}

// Message decodes and returns a message or an empty message.
func (v Value) Message() Message {
	return NewMessage(v)
}

// MessageErr decodes and returns a message or an error.
func (v Value) MessageErr() (Message, error) {
	return NewMessageErr(v)
}
