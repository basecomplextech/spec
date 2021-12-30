package spec

import (
	"fmt"
	"math"
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

	TypeFloat32 Type = 0x30
	TypeFloat64 Type = 0x31

	TypeBytes  Type = 0x40
	TypeString Type = 0x41

	TypeList    Type = 0x50
	TypeListBig Type = 0x51

	TypeMessage    Type = 0x60
	TypeMessageBig Type = 0x61
)

func CheckType(t Type) error {
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

		TypeFloat32,
		TypeFloat64,

		TypeBytes,
		TypeString,

		TypeList,
		TypeListBig,

		TypeMessage,
		TypeMessageBig:
		return nil
	}

	return fmt.Errorf("unsupported type %d", t)
}
