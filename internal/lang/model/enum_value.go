package model

import "github.com/basecomplextech/spec/internal/lang/syntax"

type EnumValue struct {
	Enum *Enum

	Name   string
	Number int
}

func parseEnumValue(enum *Enum, pval *syntax.EnumValue) (*EnumValue, error) {
	v := &EnumValue{
		Enum:   enum,
		Name:   pval.Name,
		Number: pval.Value,
	}
	return v, nil
}
