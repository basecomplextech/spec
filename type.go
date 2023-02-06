package spec

import "github.com/complex1tech/spec/encoding"

type Type = encoding.Type

const (
	TypeUndefined = encoding.TypeUndefined

	TypeTrue  = encoding.TypeTrue
	TypeFalse = encoding.TypeFalse
	TypeByte  = encoding.TypeByte

	TypeInt16 = encoding.TypeInt16
	TypeInt32 = encoding.TypeInt32
	TypeInt64 = encoding.TypeInt64

	TypeUint16 = encoding.TypeUint16
	TypeUint32 = encoding.TypeUint32
	TypeUint64 = encoding.TypeUint64

	TypeFloat32 = encoding.TypeFloat32
	TypeFloat64 = encoding.TypeFloat64

	TypeBin64  = encoding.TypeBin64
	TypeBin128 = encoding.TypeBin128
	TypeBin256 = encoding.TypeBin256

	TypeBytes  = encoding.TypeBytes
	TypeString = encoding.TypeString

	TypeList    = encoding.TypeList
	TypeBigList = encoding.TypeBigList

	TypeMessage    = encoding.TypeMessage
	TypeBigMessage = encoding.TypeBigMessage

	TypeStruct = encoding.TypeStruct
)
