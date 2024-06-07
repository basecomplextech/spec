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

	// Create imports
	for _, pimp := range pfile.Imports {
		imp, err := newImport(f, pimp)
		if err != nil {
			return nil, fmt.Errorf("%v: %w", path, err)
		}

		if err := f.addImport(imp); err != nil {
			return nil, fmt.Errorf("%v: %w", path, err)
		}
	}

	// Create options
	for _, popt := range pfile.Options {
		opt, err := newOption(popt)
		if err != nil {
			return nil, fmt.Errorf("%v: %w", path, err)
		}

		if err := f.addOption(opt); err != nil {
			return nil, fmt.Errorf("%v: %w", path, err)
		}
	}

	// Create definitions
	for _, pdef := range pfile.Definitions {
		def, err := newDefinition(pkg, f, pdef)
		if err != nil {
			return nil, err
		}

		if err := f.add(def); err != nil {
			return nil, fmt.Errorf("%v: %w", path, err)
		}
	}

	return f, nil
}

func (f *File) LookupImport(name string) (*Import, bool) {
	imp, ok := f.ImportMap[name]
	return imp, ok
}

func (f *File) lookupType(name string) (*Definition, bool) {
	def, ok := f.DefinitionNames[name]
	return def, ok
}

// internal

func (f *File) resolveImports(getPackage func(string) (*Package, error)) error {
	for _, imp := range f.Imports {
		if err := imp.resolve(getPackage); err != nil {
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

func (f *File) resolved() error {
	for _, def := range f.Definitions {
		if err := def.resolved(); err != nil {
			return fmt.Errorf("%v: %w", f.Name, err)
		}
	}
	return nil
}

func (f *File) validate() error {
	for _, def := range f.Definitions {
		if err := def.validate(); err != nil {
			return fmt.Errorf("%v: %w", f.Name, err)
		}
	}
	return nil
}

// add

func (f *File) add(def *Definition) error {
	_, ok := f.DefinitionNames[def.Name]
	if ok {
		return fmt.Errorf("duplicate definition %q", def.Name)
	}

	f.Definitions = append(f.Definitions, def)
	f.DefinitionNames[def.Name] = def
	return nil
}

func (f *File) addImport(imp *Import) error {
	_, ok := f.ImportMap[imp.Name]
	if ok {
		return fmt.Errorf("duplicate import %q", imp.Name)
	}

	f.Imports = append(f.Imports, imp)
	f.ImportMap[imp.Name] = imp
	return nil
}

func (f *File) addOption(opt *Option) error {
	_, ok := f.OptionMap[opt.Name]
	if ok {
		return fmt.Errorf("duplicate option %q", opt.Name)
	}

	f.Options = append(f.Options, opt)
	f.OptionMap[opt.Name] = opt
	return nil
}
