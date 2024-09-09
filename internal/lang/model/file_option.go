// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

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
