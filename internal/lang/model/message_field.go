// Copyright 2023 Ivan Korobkov. All rights reserved.

package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type Field struct {
	Name string
	Tag  int
	Type *Type
}

func newField(pfield *syntax.Field) (*Field, error) {
	tag := pfield.Tag
	if tag == 0 {
		return nil, fmt.Errorf("zero tag")
	}

	type_, err := newType(pfield.Type)
	if err != nil {
		return nil, err
	}

	f := &Field{
		Name: pfield.Name,
		Tag:  pfield.Tag,
		Type: type_,
	}
	return f, nil
}

func (f *Field) resolve(file *File) error {
	if err := f.Type.resolve(file); err != nil {
		return fmt.Errorf("%v: %w", f.Name, err)
	}
	return nil
}

func (f *Field) resolved() error {
	ref := f.Type.Ref
	if ref == nil {
		return nil
	}
	if ref.Type == DefinitionService {
		return fmt.Errorf("invalid field %q: service type not allowed", f.Name)
	}
	return nil
}

// Fields

type Fields struct {
	List  []*Field
	Names map[string]*Field
	Tags  map[int]*Field
}

func newFields(pfields []*syntax.Field) (*Fields, error) {
	fields := &Fields{
		Names: make(map[string]*Field),
		Tags:  make(map[int]*Field),
	}

	// Create fields
	for _, pfield := range pfields {
		field, err := newField(pfield)
		if err != nil {
			return nil, fmt.Errorf("invalid field %q: %w", pfield.Name, err)
		}

		_, ok := fields.Tags[field.Tag]
		if ok {
			return nil, fmt.Errorf("invalid field %q: duplicate tag %d", field.Name, field.Tag)
		}

		_, ok = fields.Names[field.Name]
		if ok {
			return nil, fmt.Errorf("duplicate field %q", field.Name)
		}

		fields.List = append(fields.List, field)
		fields.Names[field.Name] = field
		fields.Tags[field.Tag] = field
	}

	return fields, nil
}

func (f *Fields) Get(name string) *Field {
	return f.Names[name]
}

func (f *Fields) GetByTag(tag int) *Field {
	return f.Tags[tag]
}

// internal

func (f *Fields) resolve(file *File) error {
	for _, field := range f.List {
		if err := field.resolve(file); err != nil {
			return err
		}
	}
	return nil
}

func (f *Fields) compile() error {
	for _, field := range f.List {
		if err := field.resolved(); err != nil {
			return err
		}
	}
	return nil
}
