package ast

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
