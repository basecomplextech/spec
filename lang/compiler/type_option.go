package compiler

import "github.com/complex1tech/spec/lang/ast"

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
