package spec

import (
	"fmt"

	"github.com/complex1tech/baselibrary/types"
	"github.com/complex1tech/spec/encoding"
)

// Value is a raw value.
type Value []byte

// NewValue returns a new value from bytes.
func NewValue(b []byte) Value {
	return Value(b)
}

// ParseValue recursively parses and returns a value.
func ParseValue(b []byte) (_ Value, n int, err error) {
	t, n, err := encoding.DecodeType(b)
	if err != nil {
		return
	}

	switch t {
	default:
		err = fmt.Errorf("parse: unsupported type %d", t)
		return

	case TypeTrue, TypeFalse:
		// pass

	case TypeByte:
		_, n, err = encoding.DecodeByte(b)

	case TypeInt32:
		_, n, err = encoding.DecodeInt32(b)
	case TypeInt64:
		_, n, err = encoding.DecodeInt64(b)

	case TypeUint32:
		_, n, err = encoding.DecodeUint32(b)
	case TypeUint64:
		_, n, err = encoding.DecodeUint64(b)

	case TypeBin64:
		_, n, err = encoding.DecodeBin64(b)
	case TypeBin128:
		_, n, err = encoding.DecodeBin128(b)
	case TypeBin256:
		_, n, err = encoding.DecodeBin256(b)

	case TypeFloat32:
		_, n, err = encoding.DecodeFloat32(b)
	case TypeFloat64:
		_, n, err = encoding.DecodeFloat64(b)

	case TypeBytes:
		_, n, err = encoding.DecodeBytes(b)
	case TypeString:
		_, n, err = encoding.DecodeString(b)

	case TypeList, TypeBigList:
		_, n, err = ParseList(b)

	case TypeMessage, TypeBigMessage:
		_, n, err = ParseMessage(b)

	case TypeStruct:
		_, n, err = encoding.DecodeStruct(b)
	}

	v := Value(b)
	return v, n, nil
}

func (v Value) Type() Type {
	p, _, _ := encoding.DecodeType(v)
	return p
}

func (v Value) Bool() bool {
	p, _, _ := encoding.DecodeBool(v)
	return p
}

func (v Value) Byte() byte {
	p, _, _ := encoding.DecodeByte(v)
	return p
}

func (v Value) Int32() int32 {
	p, _, _ := encoding.DecodeInt32(v)
	return p
}

func (v Value) Int64() int64 {
	p, _, _ := encoding.DecodeInt64(v)
	return p
}

func (v Value) Uint32() uint32 {
	p, _, _ := encoding.DecodeUint32(v)
	return p
}

func (v Value) Uint64() uint64 {
	p, _, _ := encoding.DecodeUint64(v)
	return p
}

func (v Value) Float32() float32 {
	p, _, _ := encoding.DecodeFloat32(v)
	return p
}

func (v Value) Float64() float64 {
	p, _, _ := encoding.DecodeFloat64(v)
	return p
}

func (v Value) Bin64() types.Bin64 {
	p, _, _ := encoding.DecodeBin64(v)
	return p
}

func (v Value) Bin128() types.Bin128 {
	p, _, _ := encoding.DecodeBin128(v)
	return p
}

func (v Value) Bin256() types.Bin256 {
	p, _, _ := encoding.DecodeBin256(v)
	return p
}

func (v Value) Bytes() []byte {
	p, _, _ := encoding.DecodeBytes(v)
	return p
}

func (v Value) String() string {
	p, _, _ := encoding.DecodeString(v)
	return p
}

func (v Value) List() List {
	return List{}
}

func (v Value) Message() Message {
	return Message{}
}
