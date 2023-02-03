package spec

import (
	"fmt"

	"github.com/complex1tech/spec/go/encoding"
)

// Value is a raw value.
type Value []byte

// GetValue decodes and returns a value without recursive validation, or an empty value on error.
func GetValue(b []byte) Value {
	t, _, err := encoding.DecodeType(b)
	if err != nil {
		return Value{}
	}
	if err := (t).Check(); err != nil {
		return Value{}
	}
	return Value(b)
}

// DecodeValue decodes, recursively validates and returns a value.
func DecodeValue(b []byte) (_ Value, n int, err error) {
	t, n, err := encoding.DecodeType(b)
	if err != nil {
		return
	}

	switch t {
	default:
		err = fmt.Errorf("unsupported type %d", t)
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
		_, n, err = DecodeList(b, DecodeValue)

	case TypeMessage, TypeBigMessage:
		_, n, err = DecodeMessage(b)

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
