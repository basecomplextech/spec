// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

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
