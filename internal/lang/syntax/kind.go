// Copyright 2023 Ivan Korobkov. All rights reserved.

package syntax

type Kind int

const (
	KindUndefined Kind = iota
	KindAny

	KindBool
	KindByte

	KindInt16
	KindInt32
	KindInt64

	KindUint16
	KindUint32
	KindUint64

	KindFloat32
	KindFloat64

	KindBin64
	KindBin128
	KindBin256

	KindBytes
	KindString
	KindAnyMessage

	// Element-based

	KindList
	KindReference
)

// GetKind returns a type kind by its name.
func GetKind(type_ string) Kind {
	switch type_ {
	case "any":
		return KindAny

	case "bool":
		return KindBool
	case "byte":
		return KindByte

	case "int16":
		return KindInt16
	case "int32":
		return KindInt32
	case "int64":
		return KindInt64

	case "uint16":
		return KindUint16
	case "uint32":
		return KindUint32
	case "uint64":
		return KindUint64

	case "float32":
		return KindFloat32
	case "float64":
		return KindFloat64

	case "bin64":
		return KindBin64
	case "bin128":
		return KindBin128
	case "bin256":
		return KindBin256

	case "bytes":
		return KindBytes
	case "string":
		return KindString
	case "message":
		return KindAnyMessage
	}

	return KindReference
}

func (k Kind) String() string {
	switch k {
	case KindUndefined:
		return "undefined"
	case KindAny:
		return "any"

	case KindBool:
		return "bool"
	case KindByte:
		return "byte"

	case KindInt16:
		return "int16"
	case KindInt32:
		return "int32"
	case KindInt64:
		return "int64"

	case KindUint16:
		return "uint16"
	case KindUint32:
		return "uint32"
	case KindUint64:
		return "uint64"

	case KindFloat32:
		return "float32"
	case KindFloat64:
		return "float64"

	case KindBin64:
		return "bin64"
	case KindBin128:
		return "bin128"
	case KindBin256:
		return "bin256"

	case KindBytes:
		return "bytes"
	case KindString:
		return "string"
	case KindAnyMessage:
		return "message"

	case KindList:
		return "list"
	case KindReference:
		return "ref"
	}

	return "unknown"
}
