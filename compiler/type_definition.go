package compiler

import (
	"fmt"

	"github.com/complexl/spec/parser"
)

type DefinitionType string

const (
	DefinitionUndefined DefinitionType = ""
	DefinitionEnum      DefinitionType = "enum"
	DefinitionMessage   DefinitionType = "message"
	DefinitionStruct    DefinitionType = "struct"
)

func getDefinitionType(ptype parser.DefinitionType) (DefinitionType, error) {
	switch ptype {
	case parser.DefinitionEnum:
		return DefinitionEnum, nil
	case parser.DefinitionMessage:
		return DefinitionMessage, nil
	case parser.DefinitionStruct:
		return DefinitionStruct, nil
	}
	return "", fmt.Errorf("unsupported parser definition type %v", ptype)
}

// Definition

type Definition struct {
	Package *Package
	File    *File

	Name string
	Type DefinitionType

	Enum    *Enum
	Message *Message
	Struct  *Struct
}

func newDefinition(pkg *Package, file *File, pdef *parser.Definition) (*Definition, error) {
	type_, err := getDefinitionType(pdef.Type)
	if err != nil {
		return nil, err
	}

	def := &Definition{
		Package: pkg,
		File:    file,

		Name: pdef.Name,
		Type: type_,
	}

	switch type_ {
	case DefinitionEnum:
		def.Enum, err = newEnum(def, pdef.Enum)
		if err != nil {
			return nil, fmt.Errorf("%v: %w", def.Name, err)
		}

	case DefinitionMessage:
		def.Message, err = newMessage(def, pdef.Message)
		if err != nil {
			return nil, fmt.Errorf("%v: %w", def.Name, err)
		}

	case DefinitionStruct:
		def.Struct, err = newStruct(def, pdef.Struct)
		if err != nil {
			return nil, fmt.Errorf("%v: %w", def.Name, err)
		}
	}

	return def, nil
}

func (d *Definition) validate() error {
	switch d.Type {
	case DefinitionEnum:
	case DefinitionMessage:
	case DefinitionStruct:
		return d.Struct.validate()
	}
	return nil
}
