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
	TypeNil   Type = 00
	TypeTrue  Type = 01
	TypeFalse Type = 02
	TypeByte  Type = 03

	TypeInt32 Type = 10
	TypeInt64 Type = 11

	TypeUint32 Type = 20
	TypeUint64 Type = 21

	TypeU128 Type = 30
	TypeU256 Type = 31

	TypeFloat32 Type = 40
	TypeFloat64 Type = 41

	TypeBytes  Type = 50
	TypeString Type = 51

	TypeList    Type = 60
	TypeListBig Type = 61

	TypeMessage    Type = 70
	TypeMessageBig Type = 71

	TypeStruct = 80
)

func checkType(t Type) error {
	switch t {
	case
		TypeNil,
		TypeTrue,
		TypeFalse,
		TypeByte,

		TypeInt32,
		TypeInt64,

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
	case TypeByte:
		return "int8"

	case TypeInt32:
		return "int32"
	case TypeInt64:
		return "int64"

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
