package compiler

import (
	"fmt"
	"strconv"

	"github.com/complexl/spec/lang/parser"
)

type Kind int

const (
	KindUndefined Kind = iota

	// Builtin

	KindBool

	KindInt8
	KindInt16
	KindInt32
	KindInt64

	KindUint8
	KindUint16
	KindUint32
	KindUint64

	KindU128
	KindU256

	KindFloat32
	KindFloat64

	KindBytes
	KindString

	// list

	KindList

	// resolved

	KindEnum
	KindMessage
	KindStruct

	// pending

	KindReference
)

func parseKind(pkind parser.Kind) (Kind, error) {
	switch pkind {
	case parser.KindBool:
		return KindBool, nil

	case parser.KindInt8:
		return KindInt8, nil
	case parser.KindInt16:
		return KindInt16, nil
	case parser.KindInt32:
		return KindInt32, nil
	case parser.KindInt64:
		return KindInt64, nil

	case parser.KindUint8:
		return KindUint8, nil
	case parser.KindUint16:
		return KindUint16, nil
	case parser.KindUint32:
		return KindUint32, nil
	case parser.KindUint64:
		return KindUint64, nil

	case parser.KindU128:
		return KindU128, nil
	case parser.KindU256:
		return KindU256, nil

	case parser.KindFloat32:
		return KindFloat32, nil
	case parser.KindFloat64:
		return KindFloat64, nil

	case parser.KindBytes:
		return KindBytes, nil
	case parser.KindString:
		return KindString, nil

	case parser.KindList:
		return KindList, nil

	case parser.KindReference:
		return KindReference, nil
	}

	return 0, fmt.Errorf("unknown type kind %v", pkind)
}

func (k Kind) String() string {
	switch k {
	case KindBool:
		return "bool"

	case KindInt8:
		return "int8"
	case KindInt16:
		return "int16"
	case KindInt32:
		return "int32"
	case KindInt64:
		return "int64"

	case KindUint8:
		return "uint8"
	case KindUint16:
		return "uint16"
	case KindUint32:
		return "uint32"
	case KindUint64:
		return "uint64"

	case KindU128:
		return "u128"
	case KindU256:
		return "u256"

	case KindFloat32:
		return "float32"
	case KindFloat64:
		return "float64"

	case KindBytes:
		return "bytes"
	case KindString:
		return "string"

	case KindList:
		return "list"

	case KindEnum:
		return "enum"
	case KindMessage:
		return "message"
	case KindStruct:
		return "struct"

	case KindReference:
		return "reference"
	}

	return strconv.Itoa(int(k))
}
