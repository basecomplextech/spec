// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package spec

import (
	"github.com/basecomplextech/spec/internal/format"
)

// Type specifies a value type.
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
