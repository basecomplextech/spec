package model

import "github.com/basecomplextech/spec/internal/lang/ast"

type Option struct {
	Name  string
	Value string
}

func newOption(popt *ast.Option) (*Option, error) {
	opt := &Option{
		Name:  popt.Name,
		Value: popt.Value,
	}
	return opt, nil
}
