// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type Service struct {
	Package *Package
	File    *File
	Def     *Definition
	Sub     bool // Subservice

	Methods     []*Method
	MethodNames map[string]*Method
}

func newService(pkg *Package, file *File, def *Definition, ps *syntax.Service) (*Service, error) {
	srv := &Service{
		Package: pkg,
		File:    file,
		Def:     def,
		Sub:     ps.Sub,

		MethodNames: make(map[string]*Method),
	}

	if err := srv.parseMethods(ps); err != nil {
		return nil, err
	}
	return srv, nil
}

// parse

func (s *Service) parseMethods(ps *syntax.Service) error {
	for _, pm := range ps.Methods {
		if err := s.parseMethod(pm); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) parseMethod(pm *syntax.Method) error {
	method, err := parseMethod(s.Package, s.File, s, pm)
	if err != nil {
		return fmt.Errorf("%v.%v: %w", s.Def.Name, pm.Name, err)
	}

	_, ok := s.MethodNames[method.Name]
	if ok {
		return fmt.Errorf("%v.%v: duplicate method", s.Def.Name, pm.Name)
	}

	s.Methods = append(s.Methods, method)
	s.MethodNames[method.Name] = method
	return nil
}

// resolve

func (s *Service) resolve(file *File) error {
	for _, m := range s.Methods {
		if err := m.resolve(file); err != nil {
			return fmt.Errorf("%v.%v: %w", s.Def.Name, m.Name, err)
		}
	}
	return nil
}

// compile

func (s *Service) compile() error {
	for _, m := range s.Methods {
		if err := m.compile(); err != nil {
			return fmt.Errorf("%v.%v: %w", s.Def.Name, m.Name, err)
		}
	}
	return nil
}
