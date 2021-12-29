package compiler

import "github.com/baseone-run/spec/parser"

type Option struct {
	Name  string
	Value string
}

func newOption(popt *parser.Option) (*Option, error) {
	opt := &Option{
		Name:  popt.Name,
		Value: popt.Value,
	}
	return opt, nil
}
