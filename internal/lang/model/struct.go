// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package model

import (
	"fmt"

	"github.com/basecomplextech/baselibrary/collect"
	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type Struct struct {
	Package *Package
	File    *File
	Def     *Definition

	Fields collect.OrderedMap[string, *StructField]
}

func parseStruct(pkg *Package, file *File, def *Definition, ps *syntax.Struct) (*Struct, error) {
	s := &Struct{
		Package: pkg,
		File:    file,
		Def:     def,

		Fields: collect.NewOrderedMap[string, *StructField](),
	}

	if err := s.parseFields(ps); err != nil {
		return nil, err
	}
	return s, nil
}

// parse

func (s *Struct) parseFields(ps *syntax.Struct) error {
	for _, pfield := range ps.Fields {
		if err := s.parseField(pfield); err != nil {
			return err
		}
	}
	return nil
}

func (s *Struct) parseField(pfield *syntax.StructField) error {
	field, err := parseStructField(s, pfield)
	if err != nil {
		return fmt.Errorf("%v.%v: %w", s.Def.Name, pfield.Name, err)
	}

	_, ok := s.Fields.Get(field.Name)
	if ok {
		return fmt.Errorf("%v.%v: duplicate field", s.Def.Name, field.Name)
	}

	s.Fields.Put(field.Name, field)
	return nil
}

// resolve

func (s *Struct) resolve(file *File) error {
	for _, field := range s.Fields.Values() {
		if err := field.resolve(file); err != nil {
			return err
		}
	}
	return nil
}

// compile

func (s *Struct) compile() error {
	for _, field := range s.Fields.Values() {
		if err := field.compile(); err != nil {
			return err
		}
	}
	return nil
}

// validate

func (s *Struct) validate() error {
	for _, field := range s.Fields.Values() {
		if err := field.validate(); err != nil {
			return fmt.Errorf("%v: %w", s.Def.Name, err)
		}
	}
	return nil
}
