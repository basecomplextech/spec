package spec

import (
	"github.com/basecomplextech/spec/internal/core"
	"github.com/basecomplextech/spec/internal/types"
)

type Type = core.Type

const (
	TypeUndefined = core.TypeUndefined

	TypeTrue  = core.TypeTrue
	TypeFalse = core.TypeFalse
	TypeByte  = core.TypeByte

	TypeInt16 = core.TypeInt16
	TypeInt32 = core.TypeInt32
	TypeInt64 = core.TypeInt64

	TypeUint16 = core.TypeUint16
	TypeUint32 = core.TypeUint32
	TypeUint64 = core.TypeUint64

	TypeFloat32 = core.TypeFloat32
	TypeFloat64 = core.TypeFloat64

	TypeBin64  = core.TypeBin64
	TypeBin128 = core.TypeBin128
	TypeBin256 = core.TypeBin256

	TypeBytes  = core.TypeBytes
	TypeString = core.TypeString

	TypeList    = core.TypeList
	TypeBigList = core.TypeBigList

	TypeMessage    = core.TypeMessage
	TypeBigMessage = core.TypeBigMessage

	TypeStruct = core.TypeStruct
)

type (
	// Bytes is a spec byte slice backed by a buffer.
	// Clone it if you need to keep it around.
	Bytes = core.Bytes

	// String is a spec string backed by a buffer.
	// Clone it if you need to keep it around.
	String = core.String
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
