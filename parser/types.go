package parser

type File struct {
	Module *Module

	Definition  Definition
	Definitions []Definition
}

type Module struct {
	Name string

	MessageField  *MessageField
	MessageFields []*MessageField
}

type Definition interface{}

type Message struct {
	Name   string
	Fields []*MessageField
}

type MessageField struct {
	Name string
	Type string
	Tag  int
}

type Enum struct {
	Name   string
	Values []*EnumValue
}

type EnumValue struct {
	Name  string
	Value int
}
