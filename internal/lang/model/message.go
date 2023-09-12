package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/ast"
)

type Message struct {
	Package *Package
	File    *File
	Def     *Definition

	Generated bool // Auto-generated message, i.e. request/response
	Primitive bool // Message contains only primitive fields

	Fields *Fields
}

func newMessage(pkg *Package, file *File, def *Definition, pmsg *ast.Message) (*Message, error) {
	msg := &Message{
		Package: pkg,
		File:    file,
		Def:     def,
	}

	var err error
	msg.Fields, err = newFields(pmsg.Fields)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func generateMessage(pkg *Package, file *File, name string, fields *Fields) (*Message, error) {
	msg := &Message{
		Package: pkg,
		File:    file,

		Generated: true,
		Primitive: true, // Calculated below

		Fields: fields,
	}

	// Definition
	msg.Def = &Definition{
		Package: pkg,
		File:    file,

		Name:    name,
		Type:    DefinitionMessage,
		Message: msg,
	}

	// Primitive
	for _, field := range msg.Fields.List {
		ok := field.Type.builtin()
		if !ok {
			msg.Primitive = false
			break
		}
	}

	// Add to file
	if err := file.add(msg.Def); err != nil {
		return nil, err
	}
	return msg, nil
}

// internal

func (m *Message) resolve(file *File) error {
	if err := m.Fields.resolve(file); err != nil {
		return fmt.Errorf("%v.%w", m.Def.Name, err)
	}
	return nil
}

func (m *Message) resolved() error {
	if err := m.Fields.resolved(); err != nil {
		return fmt.Errorf("%v.%w", m.Def.Name, err)
	}
	return nil
}
