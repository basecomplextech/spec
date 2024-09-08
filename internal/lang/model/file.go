// Copyright 2021 Ivan Korobkov. All rights reserved.

package model

import (
	"fmt"
	"path/filepath"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type File struct {
	Package *Package
	Name    string
	Path    string

	Imports   []*Import
	ImportMap map[string]*Import // imports by names and aliases (not ids)

	Options   []*Option
	OptionMap map[string]*Option

	Definitions     []*Definition
	DefinitionNames map[string]*Definition
}

func newFile(pkg *Package, pfile *syntax.File) (*File, error) {
	path := pfile.Path
	name := filepath.Base(path)

	f := &File{
		Package: pkg,
		Name:    name,
		Path:    path,

		ImportMap:       make(map[string]*Import),
		OptionMap:       make(map[string]*Option),
		DefinitionNames: make(map[string]*Definition),
	}

	if err := f.parseImports(pfile); err != nil {
		return nil, err
	}
	if err := f.parseOptions(pfile); err != nil {
		return nil, err
	}
	if err := f.parseDefinitions(pfile); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *File) lookupImport(name string) (*Import, bool) {
	imp, ok := f.ImportMap[name]
	return imp, ok
}

func (f *File) lookupType(name string) (*Definition, bool) {
	def, ok := f.DefinitionNames[name]
	return def, ok
}

// parse imports

func (f *File) parseImports(pfile *syntax.File) error {
	for _, pimp := range pfile.Imports {
		if err := f.parseImport(pimp); err != nil {
			return err
		}
	}
	return nil
}

func (f *File) parseImport(pimp *syntax.Import) error {
	imp, err := newImport(f, pimp)
	if err != nil {
		return fmt.Errorf("%v: %w", f.Path, err)
	}

	_, ok := f.ImportMap[imp.Name]
	if ok {
		return fmt.Errorf("%v: duplicate import %q", f.Path, imp.Name)
	}

	f.Imports = append(f.Imports, imp)
	f.ImportMap[imp.Name] = imp
	return nil
}

// parse options

func (f *File) parseOptions(pfile *syntax.File) error {
	for _, popt := range pfile.Options {
		if err := f.parseOption(popt); err != nil {
			return err
		}
	}
	return nil
}

func (f *File) parseOption(popt *syntax.Option) error {
	opt, err := newOption(popt)
	if err != nil {
		return fmt.Errorf("%v: %w", f.Path, err)
	}

	_, ok := f.OptionMap[opt.Name]
	if ok {
		return fmt.Errorf("%v: duplicate option %q", f.Path, opt.Name)
	}

	f.Options = append(f.Options, opt)
	f.OptionMap[opt.Name] = opt
	return nil
}

// parse definitions

func (f *File) parseDefinitions(pfile *syntax.File) error {
	for _, pdef := range pfile.Definitions {
		if err := f.parseDefinition(pdef); err != nil {
			return err
		}
	}
	return nil
}

func (f *File) parseDefinition(pdef *syntax.Definition) error {
	def, err := parseDefinition(f.Package, f, pdef)
	if err != nil {
		return fmt.Errorf("%v: %w", f.Path, err)
	}

	return f.add(def)
}

// resolve

func (f *File) resolveImports() error {
	for _, imp := range f.Imports {
		if err := imp.resolve(); err != nil {
			return fmt.Errorf("%v: %w", f.Name, err)
		}
	}
	return nil
}

func (f *File) resolve() error {
	for _, def := range f.Definitions {
		if err := def.resolve(f); err != nil {
			return fmt.Errorf("%v: %w", f.Name, err)
		}
	}
	return nil
}

// compile

func (f *File) compile() error {
	for _, def := range f.Definitions {
		if err := def.compile(); err != nil {
			return fmt.Errorf("%v: %w", f.Path, err)
		}
	}
	return nil
}

// validate

func (f *File) validate() error {
	for _, def := range f.Definitions {
		if err := def.validate(); err != nil {
			return fmt.Errorf("%v: %w", f.Path, err)
		}
	}
	return nil
}

// add

func (f *File) add(def *Definition) error {
	_, ok := f.DefinitionNames[def.Name]
	if ok {
		return fmt.Errorf("%v: duplicate definition %q", f.Path, def.Name)
	}

	f.Definitions = append(f.Definitions, def)
	f.DefinitionNames[def.Name] = def
	return nil
}
