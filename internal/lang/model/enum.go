package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/ast"
)

type Enum struct {
	Package *Package
	File    *File
	Def     *Definition

	Values       []*EnumValue
	ValueNames   map[string]*EnumValue
	ValueNumbers map[int]*EnumValue
}

func newEnum(pkg *Package, file *File, def *Definition, penum *ast.Enum) (*Enum, error) {
	e := &Enum{
		Package: pkg,
		File:    file,
		Def:     def,

		ValueNames:   make(map[string]*EnumValue),
		ValueNumbers: make(map[int]*EnumValue),
	}

	// Create values
	for _, pval := range penum.Values {
		val, err := newEnumValue(e, pval)
		if err != nil {
			return nil, fmt.Errorf("invalid enum value %q: %w", pval.Name, err)
		}

		_, ok := e.ValueNames[val.Name]
		if ok {
			return nil, fmt.Errorf("duplicate enum value, name=%v", val.Name)
		}

		_, ok = e.ValueNumbers[val.Number]
		_, ok = e.ValueNumbers[val.Number]
		if ok {
			return nil, fmt.Errorf("duplicate enum value number, name=%v, number=%v", val.Name, val.Number)
		}

		e.Values = append(e.Values, val)
		e.ValueNames[val.Name] = val
		e.ValueNumbers[val.Number] = val
	}

	// Check zero present
	_, ok := e.ValueNumbers[0]
	if !ok {
		return nil, fmt.Errorf("zero enum value required")
	}

	return e, nil
}

// Value

type EnumValue struct {
	Enum *Enum

	Name   string
	Number int
}

func newEnumValue(enum *Enum, pval *ast.EnumValue) (*EnumValue, error) {
	v := &EnumValue{
		Enum:   enum,
		Name:   pval.Name,
		Number: pval.Value,
	}
	return v, nil
}
