package golang

import "github.com/basecomplextech/spec/internal/lang/model"

type MessageField struct {
	Package *Package
	Message *Message

	Name string
	Tag  int
	Type Type
}

func newMessageField(pkg *Package, m *Message, mf *model.Field) (*MessageField, error) {
	name := name_upperCamelCase(mf.Name)
	tag := mf.Tag

	typ, err := pkg.GetType(mf.Type)
	if err != nil {
		return nil, err
	}

	f := &MessageField{
		Package: pkg,
		Message: m,

		Name: name,
		Tag:  tag,
		Type: typ,
	}
	return f, nil
}
