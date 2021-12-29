package compiler

import (
	"fmt"

	"github.com/baseone-run/spec/parser"
)

type Struct struct {
	Def *Definition

	Fields     []*StructField
	FieldNames map[string]*StructField
}

func newStruct(def *Definition, pstr *parser.Struct) (*Struct, error) {
	str := &Struct{
		Def:        def,
		FieldNames: make(map[string]*StructField),
	}

	// create fields
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

// Field

type StructField struct {
	Name string
	Type *Type
}

func newStructField(pfield *parser.StructField) (*StructField, error) {
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
