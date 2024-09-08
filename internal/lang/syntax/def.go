// Copyright 2023 Ivan Korobkov. All rights reserved.

package syntax

type DefinitionType int

const (
	DefinitionUnknown DefinitionType = iota
	DefinitionEnum
	DefinitionMessage
	DefinitionStruct
	DefinitionService
)

type Definition struct {
	Type DefinitionType
	Name string

	Enum    *Enum
	Message *Message
	Struct  *Struct
	Service *Service
}
