package ast

type Kind int

const (
	KindUndefined Kind = iota

	// Builtin

	KindBool
	KindByte

	KindInt32
	KindInt64

	KindUint32
	KindUint64

	KindFloat32
	KindFloat64

	KindBin64
	KindBin128
	KindBin256

	KindBytes
	KindString

	// Element-based

	KindList
	KindReference
)

// GetKind returns a type kind by its name.
func GetKind(type_ string) Kind {
	switch type_ {
	case "bool":
		return KindBool
	case "byte":
		return KindByte

	case "int32":
		return KindInt32
	case "int64":
		return KindInt64

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
	}

	return KindReference
}
