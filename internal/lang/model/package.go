// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package model

import (
	"fmt"
	"path/filepath"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type Package struct {
	Context *Context

	ID   string // id is an import path as "my/example/test"
	Name string // name is "test" in "my/example/test"
	Path string // path is an absolute package path

	Files     []*File
	FileNames map[string]*File

	Options     []*Option
	OptionNames map[string]*Option

	Definitions     []*Definition
	DefinitionNames map[string]*Definition

	Compiling bool
}

func parsePackage(ctx *Context, id string, path string, pfiles []*syntax.File) (*Package, error) {
	name := filepath.Base(id)
	if name == "" || name == "." {
		return nil, fmt.Errorf("empty package name, id=%v, path=%v", id, path)
	}

	pkg := &Package{
		Context: ctx,

		ID:   id,
		Name: name,
		Path: path,

		FileNames:       make(map[string]*File),
		OptionNames:     make(map[string]*Option),
		DefinitionNames: make(map[string]*Definition),

		Compiling: true,
	}

	if err := pkg.parseFiles(pfiles); err != nil {
		return nil, err
	}
	if err := pkg.parseOptions(); err != nil {
		return nil, err
	}
	if err := pkg.parseDefinitions(); err != nil {
		return nil, err
	}
	return pkg, nil
}

func (p *Package) lookupType(name string) (*Definition, bool) {
	def, ok := p.DefinitionNames[name]
	return def, ok
}

// parse

func (p *Package) parseFiles(pfiles []*syntax.File) error {
	for _, pfile := range pfiles {
		if err := p.parseFile(pfile); err != nil {
			return err
		}
	}
	return nil
}

func (p *Package) parseFile(pfile *syntax.File) error {
	f, err := newFile(p, pfile)
	if err != nil {
		return err
	}

	p.Files = append(p.Files, f)
	p.FileNames[f.Name] = f
	return nil
}

func (p *Package) parseOptions() error {
	for _, file := range p.Files {
		for _, opt := range file.Options {
			_, ok := p.OptionNames[opt.Name]
			if ok {
				return fmt.Errorf("%v: duplicate option %q", file.Path, opt.Name)
			}

			p.Options = append(p.Options, opt)
			p.OptionNames[opt.Name] = opt
		}
	}
	return nil
}

func (p *Package) parseDefinitions() error {
	for _, file := range p.Files {
		for _, def := range file.Definitions {
			_, ok := p.DefinitionNames[def.Name]
			if ok {
				return fmt.Errorf("%v: duplicate definition %q", file.Path, def.Name)
			}

			p.Definitions = append(p.Definitions, def)
			p.DefinitionNames[def.Name] = def
		}
	}
	return nil
}

// resolve

func (p *Package) resolve() error {
	if err := p.resolveImports(); err != nil {
		return err
	}
	if err := p.resolveTypes(); err != nil {
		return err
	}
	return nil
}

func (p *Package) resolveImports() error {
	for _, file := range p.Files {
		if err := file.resolveImports(); err != nil {
			return fmt.Errorf("%v: %w", p.Name, err)
		}
	}
	return nil
}

func (p *Package) resolveTypes() error {
	for _, file := range p.Files {
		if err := file.resolve(); err != nil {
			return fmt.Errorf("%v: %w", p.Name, err)
		}
	}
	return nil
}

// compile

func (p *Package) compile() error {
	for _, file := range p.Files {
		if err := file.compile(); err != nil {
			return fmt.Errorf("%v: %w", p.Name, err)
		}
	}
	return nil
}

// validate

func (p *Package) validate() error {
	for _, file := range p.Files {
		if err := file.validate(); err != nil {
			return fmt.Errorf("%v: %w", p.Name, err)
		}
	}
	return nil
}
