package golang

import "github.com/basecomplextech/spec/internal/lang/model"

type Message struct {
	Package *Package

	Name   string
	Fields []*MessageField
}

func newMessage(pkg *Package, mdef *model.Definition) (*Message, error) {
	m := &Message{
		Package: pkg,
		Name:    mdef.Name,
	}

	for _, mf := range mdef.Message.Fields.List {
		f, err := newMessageField(pkg, m, mf)
		if err != nil {
			return nil, err
		}
		m.Fields = append(m.Fields, f)
	}
	return nil, nil
}
