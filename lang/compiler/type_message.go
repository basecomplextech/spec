package compiler

import (
	"fmt"

	"github.com/complex1tech/spec/lang/ast"
)

type Message struct {
	Def *Definition

	Fields     []*MessageField
	FieldTags  map[int]*MessageField
	FieldNames map[string]*MessageField
}

func newMessage(def *Definition, pmsg *ast.Message) (*Message, error) {
	msg := &Message{
		Def: def,

		FieldNames: make(map[string]*MessageField),
		FieldTags:  make(map[int]*MessageField),
	}

	// Create fields
	for _, pfield := range pmsg.Fields {
		field, err := newMessageField(pfield)
		if err != nil {
			return nil, fmt.Errorf("invalid field %q: %w", pfield.Name, err)
		}

		_, ok := msg.FieldTags[field.Tag]
		if ok {
			return nil, fmt.Errorf("duplicate field tag, name=%v, tag=%d",
				field.Name, field.Tag)
		}

		_, ok = msg.FieldNames[field.Name]
		if ok {
			return nil, fmt.Errorf("duplicate field name, name=%v", field.Name)
		}

		msg.Fields = append(msg.Fields, field)
		msg.FieldNames[field.Name] = field
		msg.FieldTags[field.Tag] = field
	}

	return msg, nil
}

// Field

type MessageField struct {
	Name string
	Tag  int
	Type *Type
}

func newMessageField(pfield *ast.MessageField) (*MessageField, error) {
	tag := pfield.Tag
	if tag == 0 {
		return nil, fmt.Errorf("zero tag")
	}

	type_, err := newType(pfield.Type)
	if err != nil {
		return nil, err
	}

	f := &MessageField{
		Name: pfield.Name,
		Tag:  pfield.Tag,
		Type: type_,
	}
	return f, nil
}
