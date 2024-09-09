// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package core

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

	TypeInt16 Type = 10
	TypeInt32 Type = 11
	TypeInt64 Type = 12

	TypeUint16 Type = 20
	TypeUint32 Type = 21
	TypeUint64 Type = 22

	TypeFloat32 Type = 40
	TypeFloat64 Type = 41

	TypeBin64  Type = 30
	TypeBin128 Type = 31
	TypeBin256 Type = 32

	TypeBytes  Type = 50
	TypeString Type = 60

	TypeList    Type = 70
	TypeBigList Type = 71

	TypeMessage    Type = 80
	TypeBigMessage Type = 81

	TypeStruct Type = 90
)

func (t Type) Check() error {
	switch t {
	case
		TypeTrue,
		TypeFalse,
		TypeByte,

		TypeInt16,
		TypeInt32,
		TypeInt64,

		TypeUint16,
		TypeUint32,
		TypeUint64,

		TypeFloat32,
		TypeFloat64,

		TypeBin64,
		TypeBin128,
		TypeBin256,

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

	case TypeInt16:
		return "int16"
	case TypeInt32:
		return "int32"
	case TypeInt64:
		return "int64"

	case TypeUint16:
		return "uint16"
	case TypeUint32:
		return "uint32"
	case TypeUint64:
		return "uint64"

	case TypeFloat32:
		return "float32"
	case TypeFloat64:
		return "float64"

	case TypeBin64:
		return "bin64"
	case TypeBin128:
		return "bin128"
	case TypeBin256:
		return "bin256"

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
