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
	TypeUndefined Type = 0

	TypeTrue  Type = 01
	TypeFalse Type = 02
	TypeByte  Type = 03

	TypeInt32 Type = 10
	TypeInt64 Type = 11

	TypeUint32 Type = 20
	TypeUint64 Type = 21

	TypeB128 Type = 30
	TypeB256 Type = 31

	TypeFloat32 Type = 40
	TypeFloat64 Type = 41

	TypeBytes  Type = 50
	TypeString Type = 60

	TypeList    Type = 70
	TypeBigList Type = 71

	TypeMessage    Type = 80
	TypeBigMessage Type = 81

	TypeStruct Type = 90
)

func checkType(t Type) error {
	switch t {
	case
		TypeTrue,
		TypeFalse,
		TypeByte,

		TypeInt32,
		TypeInt64,

		TypeUint32,
		TypeUint64,

		TypeB128,
		TypeB256,

		TypeFloat32,
		TypeFloat64,

		TypeBytes,
		TypeString,

		TypeList,
		TypeBigList,

		TypeMessage,
		TypeBigMessage,

		TypeStruct:
		return nil
	}

	return fmt.Errorf("unsupported type %d", t)
}

func (t Type) String() string {
	switch t {
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

	case TypeB128:
		return "bin128"
	case TypeB256:
		return "bin256"

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
	case TypeBigList:
		return "big_list"

	case TypeMessage:
		return "message"
	case TypeBigMessage:
		return "big_message"

	case TypeStruct:
		return "struct"
	}

	return strconv.Itoa(int(t))
}
