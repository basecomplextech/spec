package model

import (
	"fmt"

	"github.com/basecomplextech/baselibrary/collect/orderedmap"
	"github.com/basecomplextech/spec/internal/lang/ast"
)

type Struct struct {
	Def *Definition

	Fields *orderedmap.Map[string, *StructField]
}

func newStruct(def *Definition, pstr *ast.Struct) (*Struct, error) {
	str := &Struct{
		Def:    def,
		Fields: orderedmap.New[string, *StructField](),
	}

	// Create fields
	for _, pfield := range pstr.Fields {
		field, err := newStructField(pfield)
		if err != nil {
			return nil, fmt.Errorf("invalid field %q: %w", pfield.Name, err)
		}

		_, ok := str.Fields.Get(field.Name)
		if ok {
			return nil, fmt.Errorf("duplicate field, name=%v", field.Name)
		}

		str.Fields.Put(field.Name, field)
	}

	return str, nil
}

func (s *Struct) resolve(file *File) error {
	for _, field := range s.Fields.Values() {
		if err := field.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", s.Def.Name, err)
		}
	}
	return nil
}

func (s *Struct) resolved() error {
	for _, field := range s.Fields.Values() {
		if err := field.resolved(); err != nil {
			return fmt.Errorf("%v: %w", s.Def.Name, err)
		}
	}
	return nil
}

func (s *Struct) validate() error {
	for _, field := range s.Fields.Values() {
		if err := field.validate(); err != nil {
			return fmt.Errorf("%v: %w", s.Def.Name, err)
		}
	}
	return nil
}

// Field

type StructField struct {
	Name string
	Type *Type
}

func newStructField(pfield *ast.StructField) (*StructField, error) {
	type_, err := newType(pfield.Type)
	if err != nil {
		return nil, err
	}

	f := &StructField{
		Name: pfield.Name,
		Type: type_,
	}
	return f, nil
}

func (f *StructField) resolve(file *File) error {
	if err := f.Type.resolve(file); err != nil {
		return fmt.Errorf("%v: %w", f.Name, err)
	}
	return nil
}

func (f *StructField) resolved() error {
	ref := f.Type.Ref
	if ref == nil {
		return nil
	}
	if ref.Type == DefinitionService {
		return fmt.Errorf("%v: service type not allowed", f.Name)
	}
	return nil
}

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
