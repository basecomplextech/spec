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
type Type byte

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

	TypeBytes    Type = 50
	TypeBytesBig Type = 51

	TypeString    Type = 60
	TypeBigString Type = 61

	TypeList    Type = 70
	TypeBigList Type = 71

	TypeMessage    Type = 80
	TypeBigMessage Type = 81

	TypeStruct    Type = 90
	TypeBigStruct Type = 91
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
		TypeBytesBig,

		TypeString,
		TypeBigString,

		TypeList,
		TypeBigList,

		TypeMessage,
		TypeBigMessage,

		TypeStruct,
		TypeBigStruct:
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
	case TypeBytesBig:
		return "big_bytes"

	case TypeString:
		return "string"
	case TypeBigString:
		return "big_string"

	case TypeList:
		return "list"
	case TypeBigList:
		return "big_list"

	case TypeMessage:
		return "message"
	case TypeBigMessage:
		return "big_message"

	case TypeStruct:
		return "struct"
	case TypeBigStruct:
		return "big_struct"
	}

	return strconv.Itoa(int(t))
}
