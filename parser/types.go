package parser

type File struct {
	Imports     []*Import
	Definitions []Definition
}

// Import

type Import struct {
	ID    string
	Alias string
}

// Definitions

type Definition interface{}

// Enum

type Enum struct {
	Name   string
	Values []*EnumValue
}

type EnumValue struct {
	Name  string
	Value int
}

// Message

type Message struct {
	Name   string
	Fields []*MessageField
}

type MessageField struct {
	Name string
	Type *Type
	Tag  int
}

// Struct

type Struct struct {
	Name   string
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
