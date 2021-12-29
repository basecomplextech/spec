package parser

type File struct {
	Path        string
	Imports     []*Import
	Definitions []*Definition
}

// Import

type Import struct {
	ID    string
	Alias string
}

// Definitions

type DefinitionType int

const (
	DefinitionUnknown DefinitionType = iota
	DefinitionEnum
	DefinitionMessage
	DefinitionStruct
)

type Definition struct {
	Type DefinitionType
	Name string

	Enum    *Enum
	Message *Message
	Struct  *Struct
}

// Enum

type Enum struct {
	Values []*EnumValue
}

type EnumValue struct {
	Name  string
	Value int
}

// Message

type Message struct {
	Fields []*MessageField
}

type MessageField struct {
	Name string
	Type *Type
	Tag  int
}

// Struct

type Struct struct {
	Fields []*StructField
}

type StructField struct {
	Name string
	Type *Type
}

// Type

type Type struct {
	Kind    Kind
	Name    string
	Import  string // package name in imported type, "pkg" in "pkg.Name"
	Element *Type  // element type in list and nullable types
}

// Kind

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

	KindFloat32
	KindFloat64

	KindBytes
	KindString

	// Element-based

	KindList
	KindReference
)

// getKind returns a type kind by its name.
func getKind(type_ string) Kind {
	switch type_ {
	case "bool":
		return KindBool

	case "int8":
		return KindInt8
	case "int16":
		return KindInt16
	case "int32":
		return KindInt32
	case "int64":
		return KindInt64

	case "uint8":
		return KindUint8
	case "uint16":
		return KindUint16
	case "uint32":
		return KindUint32
	case "uint64":
		return KindUint64

	case "float32":
		return KindUint32
	case "float64":
		return KindUint64

	case "bytes":
		return KindBytes
	case "string":
		return KindString
	}

	return KindReference
}
