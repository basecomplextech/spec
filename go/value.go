package spec

import (
	"fmt"
)

// Value is a raw value.
type Value []byte

// GetValue decodes and returns a value without recursive validation, or an empty value on error.
func GetValue(b []byte) Value {
	t, _, err := DecodeType(b)
	if err != nil {
		return Value{}
	}
	if err := checkType(t); err != nil {
		return Value{}
	}
	return Value(b)
}

// DecodeValue decodes, recursively validates and returns a value.
func DecodeValue(b []byte) (_ Value, n int, err error) {
	t, n, err := DecodeType(b)
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
		_, n, err = DecodeByte(b)

	case TypeInt32:
		_, n, err = DecodeInt32(b)
	case TypeInt64:
		_, n, err = DecodeInt64(b)

	case TypeUint32:
		_, n, err = DecodeUint32(b)
	case TypeUint64:
		_, n, err = DecodeUint64(b)

	case TypeBin64:
		_, n, err = DecodeBin64(b)
	case TypeBin128:
		_, n, err = DecodeBin128(b)
	case TypeBin256:
		_, n, err = DecodeBin256(b)

	case TypeFloat32:
		_, n, err = DecodeFloat32(b)
	case TypeFloat64:
		_, n, err = DecodeFloat64(b)

	case TypeBytes:
		_, n, err = DecodeBytes(b)
	case TypeString:
		_, n, err = DecodeString(b)

	case TypeList, TypeBigList:
		_, n, err = DecodeList(b, DecodeValue)

	case TypeMessage, TypeBigMessage:
		_, n, err = DecodeMessage(b)

	case TypeStruct:
		_, n, err = DecodeStruct(b)
	}

	v := Value(b)
	return v, n, nil
}

func (v Value) Type() Type {
	p, _, _ := DecodeType(v)
	return p
}
