package model

import (
	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type Message struct {
	Package *Package
	File    *File
	Def     *Definition

	Fields    *Fields
	Generated bool // Auto-generated message, i.e. request/response
}

func parseMessage(pkg *Package, file *File, def *Definition, pmsg *syntax.Message) (*Message, error) {
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

func generateMessageDef(pkg *Package, file *File, name string, fields *Fields) (*Definition, error) {
	def := &Definition{
		Package: pkg,
		File:    file,

		Name: name,
		Type: DefinitionMessage,
	}

	// Generate message
	msg, err := generateMessage(pkg, file, def, fields)
	if err != nil {
		return nil, err
	}
	def.Message = msg

	// Add definition to file
	if err := file.add(msg.Def); err != nil {
		return nil, err
	}
	return def, nil
}

func generateMessage(pkg *Package, file *File, def *Definition, fields *Fields) (*Message, error) {
	msg := &Message{
		Package: pkg,
		File:    file,
		Def:     def,

		Fields:    fields,
		Generated: true,
	}
	return msg, nil
}

// resolve

func (m *Message) resolve(file *File) error {
	return m.Fields.resolve(file)
}

// compile

func (m *Message) compile() error {
	return m.Fields.compile()
}

// validate

func (m *Message) validate() error {
	return nil
}
