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

type Kind int

const (
	KindUndefined Kind = iota
	KindBase
	KindList
	KindNullable
)

type Type struct {
	Kind  Kind
	Ident string
}
