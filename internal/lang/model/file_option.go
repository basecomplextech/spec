// Copyright 2021 Ivan Korobkov. All rights reserved.

package model

import "github.com/basecomplextech/spec/internal/lang/syntax"

type Option struct {
	Name  string
	Value string
}

func newOption(popt *syntax.Option) (*Option, error) {
	opt := &Option{
		Name:  popt.Name,
		Value: popt.Value,
	}
	return opt, nil
}
