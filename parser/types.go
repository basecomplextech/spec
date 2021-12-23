package parser

type File struct {
	Module      string
	Import      *Import
	Definitions []Definition
}

// Import

type Import struct {
	Modules []*ImportModule
}

type ImportModule struct {
	Name  string
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
	Type string
	Tag  int
}

// Struct

type Struct struct {
	Name   string
	Fields []*StructField
}

type StructField struct {
	Name string
	Type string
}
