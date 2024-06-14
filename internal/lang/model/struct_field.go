package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type StructField struct {
	Struct *Struct
	Name   string
	Type   *Type
}

func parseStructField(str *Struct, pfield *syntax.StructField) (*StructField, error) {
	typ, err := newType(pfield.Type)
	if err != nil {
		return nil, err
	}

	f := &StructField{
		Struct: str,
		Name:   pfield.Name,
		Type:   typ,
	}
	return f, nil
}

// resolve

func (f *StructField) resolve(file *File) error {
	if err := f.Type.resolve(file); err != nil {
		return fmt.Errorf("%v: %w", f.Name, err)
	}
	return nil
}

// compile

func (f *StructField) compile() error {
	ref := f.Type.Ref
	if ref == nil {
		return nil
	}
	if ref.Type == DefinitionService {
		return fmt.Errorf("%v: service type not allowed", f.Name)
	}
	return nil
}

// validate

func (f *StructField) validate() error {
	t := f.Type

	switch {
	case t.builtin():
		return nil
	case t.Kind == KindStruct:
		return nil
	}

	return fmt.Errorf("%v: structs support only value types or other structs, actual=%v",
		f.Name, t.Kind)
}
