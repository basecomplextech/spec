package compiler

import "github.com/baseone-run/spec/parser"

type Type struct {
}

func newType(ptype *parser.Type) (*Type, error) {
	type_ := &Type{}
	return type_, nil
}
