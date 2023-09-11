package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/ast"
)

type Service struct {
	Def *Definition
	Sub bool // Subservice

	Methods     []*Method
	MethodNames map[string]*Method
}

func newService(def *Definition, psrv *ast.Service) (*Service, error) {
	srv := &Service{
		Def: def,
		Sub: psrv.Sub,

		MethodNames: make(map[string]*Method),
	}

	// Create methods
	for _, pm := range psrv.Methods {
		method, err := newMethod(pm)
		if err != nil {
			return nil, fmt.Errorf("%v: %w", pm.Name, err)
		}

		_, ok := srv.MethodNames[method.Name]
		if ok {
			return nil, fmt.Errorf("%v: duplicate method", pm.Name)
		}

		srv.Methods = append(srv.Methods, method)
		srv.MethodNames[method.Name] = method
	}

	return srv, nil
}

func (s *Service) resolve(file *File) error {
	for _, method := range s.Methods {
		if err := method.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", s.Def.Name, err)
		}
	}
	return nil
}

func (s *Service) resolved() error {
	for _, m := range s.Methods {
		if err := m.resolved(); err != nil {
			return fmt.Errorf("%v: %w", s.Def.Name, err)
		}
	}
	return nil
}
