// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type Enum struct {
	Package *Package
	File    *File
	Def     *Definition

	Values       []*EnumValue
	ValueNames   map[string]*EnumValue
	ValueNumbers map[int]*EnumValue
}

func parseEnum(pkg *Package, file *File, def *Definition, penum *syntax.Enum) (*Enum, error) {
	e := &Enum{
		Package: pkg,
		File:    file,
		Def:     def,

		ValueNames:   make(map[string]*EnumValue),
		ValueNumbers: make(map[int]*EnumValue),
	}

	// Parse values
	if err := e.parseValues(penum); err != nil {
		return nil, err
	}

	// Check zero
	_, ok := e.ValueNumbers[0]
	if !ok {
		return nil, fmt.Errorf("zero enum value required")
	}
	return e, nil
}

// parse

func (e *Enum) parseValues(penum *syntax.Enum) error {
	for _, pval := range penum.Values {
		if err := e.parseValue(pval); err != nil {
			return err
		}
	}
	return nil
}

func (e *Enum) parseValue(pval *syntax.EnumValue) error {
	val, err := parseEnumValue(e, pval)
	if err != nil {
		return fmt.Errorf("%v.%v: %w", e.Def.Name, pval.Name, err)
	}

	// Check name
	_, ok := e.ValueNames[val.Name]
	if ok {
		return fmt.Errorf("%v.%v: duplicate enum value", e.Def.Name, val.Name)
	}

	// Check number
	_, ok = e.ValueNumbers[val.Number]
	if ok {
		return fmt.Errorf("%v.%v: duplicate enum value number, number=%v",
			e.Def.Name, val.Name, val.Number)
	}

	// Add value
	e.Values = append(e.Values, val)
	e.ValueNames[val.Name] = val
	e.ValueNumbers[val.Number] = val
	return nil
}
