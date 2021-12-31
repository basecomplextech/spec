package spec

import (
	"fmt"
	"math"
	"strconv"
)

const (
	MaxSize = math.MaxInt32
)

// Type specifies a value type.
type Type uint8

const (
	TypeNil   Type = 0x00
	TypeTrue  Type = 0x01
	TypeFalse Type = 0x02

	TypeInt8  Type = 0x10
	TypeInt16 Type = 0x11
	TypeInt32 Type = 0x12
	TypeInt64 Type = 0x13

	TypeUint8  Type = 0x20
	TypeUint16 Type = 0x21
	TypeUint32 Type = 0x22
	TypeUint64 Type = 0x23

	TypeU128 Type = 0x24
	TypeU256 Type = 0x25

	TypeFloat32 Type = 0x30
	TypeFloat64 Type = 0x31

	TypeBytes  Type = 0x40
	TypeString Type = 0x41

	TypeList    Type = 0x50
	TypeListBig Type = 0x51

	TypeMessage    Type = 0x60
	TypeMessageBig Type = 0x61

	TypeStruct = 0x70
)

func checkType(t Type) error {
	switch t {
	case
		TypeNil,
		TypeTrue,
		TypeFalse,

		TypeInt8,
		TypeInt16,
		TypeInt32,
		TypeInt64,

		TypeUint8,
		TypeUint16,
		TypeUint32,
		TypeUint64,

		TypeU128,
		TypeU256,

		TypeFloat32,
		TypeFloat64,

		TypeBytes,
		TypeString,

		TypeList,
		TypeListBig,

		TypeMessage,
		TypeMessageBig,

		TypeStruct:
		return nil
	}

	return fmt.Errorf("unsupported type %d", t)
}

func (t Type) String() string {
	switch t {
	case TypeNil:
		return "nil"
	case TypeTrue:
		return "true"
	case TypeFalse:
		return "false"

	case TypeInt8:
		return "int8"
	case TypeInt16:
		return "int16"
	case TypeInt32:
		return "int32"
	case TypeInt64:
		return "int64"

	case TypeUint8:
		return "uint8"
	case TypeUint16:
		return "uint16"
	case TypeUint32:
		return "uint32"
	case TypeUint64:
		return "uint64"

	case TypeU128:
		return "u128"
	case TypeU256:
		return "u256"

	case TypeFloat32:
		return "float32"
	case TypeFloat64:
		return "float64"

	case TypeBytes:
		return "bytes"
	case TypeString:
		return "string"

	case TypeList:
		return "list"
	case TypeListBig:
		return "list_big"

	case TypeMessage:
		return "message"
	case TypeMessageBig:
		return "message_big"

	case TypeStruct:
		return "struct"
	}

	return strconv.Itoa(int(t))
}
