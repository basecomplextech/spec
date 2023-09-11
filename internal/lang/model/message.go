package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/ast"
)

type Message struct {
	Def *Definition

	Fields *Fields
}

func newMessage(def *Definition, pmsg *ast.Message) (*Message, error) {
	msg := &Message{
		Def: def,
	}

	var err error
	msg.Fields, err = newFields(pmsg.Fields)
	if err != nil {
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
