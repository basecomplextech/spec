// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package spec

import (
	"github.com/basecomplextech/spec/internal/format"
	"github.com/basecomplextech/spec/internal/types"
)

type Type = format.Type

const (
	TypeUndefined = format.TypeUndefined

	TypeTrue  = format.TypeTrue
	TypeFalse = format.TypeFalse
	TypeByte  = format.TypeByte

	TypeInt16 = format.TypeInt16
	TypeInt32 = format.TypeInt32
	TypeInt64 = format.TypeInt64

	TypeUint16 = format.TypeUint16
	TypeUint32 = format.TypeUint32
	TypeUint64 = format.TypeUint64

	TypeFloat32 = format.TypeFloat32
	TypeFloat64 = format.TypeFloat64

	TypeBin64  = format.TypeBin64
	TypeBin128 = format.TypeBin128
	TypeBin256 = format.TypeBin256

	TypeBytes  = format.TypeBytes
	TypeString = format.TypeString

	TypeList    = format.TypeList
	TypeBigList = format.TypeBigList

	TypeMessage    = format.TypeMessage
	TypeBigMessage = format.TypeBigMessage

	TypeStruct = format.TypeStruct
)

type (
	// Bytes is a spec byte slice backed by a buffer.
	// Clone it if you need to keep it around.
	Bytes = format.Bytes

	// String is a spec string backed by a buffer.
	// Clone it if you need to keep it around.
	String = format.String
)

type (
	// List is a raw list of elements.
	List = types.List

	// Message is a raw message.
	Message = types.Message

	// MessageType is a type implemented by generated messages.
	MessageType = types.MessageType

	// Value is a raw value.
	Value = types.Value
)

// List

// NewList returns a new list from bytes or an empty list when not a list.
func NewList(b []byte) List {
	return types.NewList(b)
}

// NewListErr returns a new list from bytes or an error when not a list.
func NewListErr(b []byte) (List, error) {
	return types.NewListErr(b)
}

// ParseList recursively parses and returns a list.
func ParseList(b []byte) (l List, size int, err error) {
	return types.ParseList(b)
}

// Message

// NewMessage returns a new message from bytes or an empty message when not a message.
func NewMessage(b []byte) Message {
	return types.NewMessage(b)
}

// NewMessageErr returns a new message from bytes or an error when not a message.
func NewMessageErr(b []byte) (Message, error) {
	return types.NewMessageErr(b)
}

// ParseMessage recursively parses and returns a message.
func ParseMessage(b []byte) (_ Message, size int, err error) {
	return types.ParseMessage(b)
}

// Value

// NewValue returns a new value from bytes or nil when not valid.
func NewValue(b []byte) Value {
	return types.NewValue(b)
}

// NewValueErr returns a new value from bytes or an error when not valid.
func NewValueErr(b []byte) (Value, error) {
	return types.NewValueErr(b)
}

// ParseValue recursively parses and returns a value.
func ParseValue(b []byte) (_ Value, n int, err error) {
	return types.ParseValue(b)
}
