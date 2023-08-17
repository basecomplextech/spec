package compiler

import (
	"fmt"

	"github.com/basecomplextech/spec/lang/ast"
)

type Service struct {
	Def *Definition

	Methods     []*Method
	MethodNames map[string]*Method
}

func newService(def *Definition, psrv *ast.Service) (*Service, error) {
	srv := &Service{
		Def: def,

		MethodNames: make(map[string]*Method),
	}

	// Create methods
	for _, pm := range psrv.Methods {
		method, err := newMethod(pm)
		if err != nil {
			return nil, fmt.Errorf("invalid method %q: %w", pm.Name, err)
		}

		_, ok := srv.MethodNames[method.Name]
		if ok {
			return nil, fmt.Errorf("duplicate method %q", pm.Name)
		}

		srv.Methods = append(srv.Methods, method)
		srv.MethodNames[method.Name] = method
	}

	return srv, nil
}

// Method

type Method struct {
	Name string

	Args     []*MethodArg
	ArgNames map[string]*MethodArg

	Results     []*MethodResult
	ResultNames map[string]*MethodResult
}

func newMethod(pm *ast.Method) (*Method, error) {
	m := &Method{
		Name: pm.Name,

		ArgNames:    make(map[string]*MethodArg),
		ResultNames: make(map[string]*MethodResult),
	}

	// Create args
	for _, parg := range pm.Args {
		arg, err := newMethodArg(parg)
		if err != nil {
			return nil, fmt.Errorf("invalid arg %q: %w", arg.Name, err)
		}

		_, ok := m.ArgNames[arg.Name]
		if ok {
			return nil, fmt.Errorf("duplicate arg %q", arg.Name)
		}

		m.Args = append(m.Args, arg)
		m.ArgNames[arg.Name] = arg
	}

	// Create results
	for _, pr := range pm.Results {
		result, err := newMethodResult(pr)
		if err != nil {
			return nil, fmt.Errorf("invalid result %q", result.Name)
		}

		_, ok := m.ResultNames[result.Name]
		if ok {
			return nil, fmt.Errorf("duplicate result %q", result.Name)
		}

		m.Results = append(m.Results, result)
		m.ResultNames[result.Name] = result
	}

	return m, nil
}

// Arg

type MethodArg struct {
	Name string
	Type *Type
}

func newMethodArg(p *ast.MethodArg) (*MethodArg, error) {
	type_, err := newType(p.Type)
	if err != nil {
		return nil, err
	}

	a := &MethodArg{
		Name: p.Name,
		Type: type_,
	}
	return a, nil
}

// Result

type MethodResult struct {
	Name string
	Type *Type
}

func newMethodResult(p *ast.MethodResult) (*MethodResult, error) {
	type_, err := newType(p.Type)
	if err != nil {
		return nil, err
	}

	r := &MethodResult{
		Name: p.Name,
		Type: type_,
	}
	return r, nil
}
