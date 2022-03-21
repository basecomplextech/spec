package parser

type File struct {
	Path        string
	Imports     []*Import
	Options     []*Option
	Definitions []*Definition
}

// Import

type Import struct {
	ID    string
	Alias string
}

// Option

type Option struct {
	Name  string
	Value string
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
	KindByte

	KindInt32
	KindInt64
	KindUint32
	KindUint64

	KindU128
	KindU256

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

	case "u128":
		return KindU128
	case "u256":
		return KindU256

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
