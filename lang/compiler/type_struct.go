package compiler

import (
	"fmt"

	"github.com/basecomplextech/spec/lang/ast"
)

type Struct struct {
	Def *Definition

	Fields     []*StructField
	FieldNames map[string]*StructField
}

func newStruct(def *Definition, pstr *ast.Struct) (*Struct, error) {
	str := &Struct{
		Def:        def,
		FieldNames: make(map[string]*StructField),
	}

	// Create fields
	for _, pfield := range pstr.Fields {
		field, err := newStructField(pfield)
		if err != nil {
			return nil, fmt.Errorf("invalid field %q: %w", pfield.Name, err)
		}

		_, ok := str.FieldNames[field.Name]
		if ok {
			return nil, fmt.Errorf("duplicate field, name=%v", field.Name)
		}

		str.Fields = append(str.Fields, field)
		str.FieldNames[field.Name] = field
	}

	return str, nil
}

func (s *Struct) validate() error {
	for _, field := range s.Fields {
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
