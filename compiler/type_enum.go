package compiler

import (
	"fmt"

	"github.com/baseone-run/spec/parser"
)

type Enum struct {
	Values       []*EnumValue
	ValueNames   map[string]*EnumValue
	ValueNumbers map[int]*EnumValue
}

func newEnum(penum *parser.Enum) (*Enum, error) {
	e := &Enum{
		ValueNames:   make(map[string]*EnumValue),
		ValueNumbers: make(map[int]*EnumValue),
	}

	// create values
	for _, pval := range penum.Values {
		val, err := newEnumValue(pval)
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

	// check zero present
	_, ok := e.ValueNumbers[0]
	if !ok {
		return nil, fmt.Errorf("zero enum value required")
	}

	return e, nil
}

// Value

type EnumValue struct {
	Name   string
	Number int
}

func newEnumValue(pval *parser.EnumValue) (*EnumValue, error) {
	v := &EnumValue{
		Name:   pval.Name,
		Number: pval.Value,
	}
	return v, nil
}
