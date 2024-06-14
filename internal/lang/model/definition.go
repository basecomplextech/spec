package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type Definition struct {
	Package *Package
	File    *File

	Name string
	Type DefinitionType

	Enum    *Enum
	Message *Message
	Struct  *Struct
	Service *Service
}

func parseDefinition(pkg *Package, file *File, pdef *syntax.Definition) (*Definition, error) {
	typ, err := parseDefinitionType(pdef.Type)
	if err != nil {
		return nil, err
	}

	def := &Definition{
		Package: pkg,
		File:    file,

		Name: pdef.Name,
		Type: typ,
	}

	if err := def.parse(pdef); err != nil {
		return nil, err
	}
	return def, nil
}

// parse

func (d *Definition) parse(pdef *syntax.Definition) (err error) {
	switch d.Type {
	case DefinitionEnum:
		d.Enum, err = parseEnum(d.Package, d.File, d, pdef.Enum)
		return err

	case DefinitionMessage:
		d.Message, err = parseMessage(d.Package, d.File, d, pdef.Message)
		return err

	case DefinitionStruct:
		d.Struct, err = parseStruct(d.Package, d.File, d, pdef.Struct)
		return err

	case DefinitionService:
		d.Service, err = newService(d.Package, d.File, d, pdef.Service)
		return err
	}

	panic(fmt.Sprintf("unsupported definition type %q", d.Type))
}

// resolve

func (d *Definition) resolve(file *File) error {
	switch d.Type {
	case DefinitionEnum:
		return nil
	case DefinitionMessage:
		return d.Message.resolve(file)
	case DefinitionStruct:
		return d.Struct.resolve(file)
	case DefinitionService:
		return d.Service.resolve(file)
	}
	return nil
}

// compile

func (d *Definition) compile() error {
	switch d.Type {
	case DefinitionEnum:
		return nil
	case DefinitionMessage:
		return d.Message.compile()
	case DefinitionStruct:
		return d.Struct.compile()
	case DefinitionService:
		return d.Service.compile()
	}
	return nil
}

// validate

func (d *Definition) validate() error {
	switch d.Type {
	case DefinitionEnum:
		return nil
	case DefinitionMessage:
		return d.Message.validate()
	case DefinitionStruct:
		return d.Struct.validate()
	}
	return nil
}
