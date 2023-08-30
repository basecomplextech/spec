package model

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/basecomplextech/spec/internal/lang/ast"
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

func newFile(pkg *Package, pfile *ast.File) (*File, error) {
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

		_, ok := f.ImportMap[imp.Name]
		if ok {
			return nil, fmt.Errorf("%v: duplicate import %v", path, imp.Name)
		}

		f.Imports = append(f.Imports, imp)
		f.ImportMap[imp.Name] = imp
	}

	// Create options
	for _, popt := range pfile.Options {
		opt, err := newOption(popt)
		if err != nil {
			return nil, fmt.Errorf("%v: %w", path, err)
		}

		_, ok := f.ImportMap[opt.Name]
		if ok {
			return nil, fmt.Errorf("%v: duplicate option %v", path, opt.Name)
		}

		f.Options = append(f.Options, opt)
		f.OptionMap[opt.Name] = opt
	}

	// Create definitions
	for _, pdef := range pfile.Definitions {
		def, err := newDefinition(pkg, f, pdef)
		if err != nil {
			return nil, err
		}

		_, ok := f.DefinitionNames[def.Name]
		if ok {
			return nil, fmt.Errorf("duplicate definition in file, path=%v", path)
		}

		f.Definitions = append(f.Definitions, def)
		f.DefinitionNames[def.Name] = def
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

func (f *File) validate() error {
	for _, def := range f.Definitions {
		if err := def.validate(); err != nil {
			return fmt.Errorf("%v: %w", f.Name, err)
		}
	}
	return nil
}

// Import

type Import struct {
	File *File

	ID      string   // full id
	Name    string   // name or alias
	Package *Package // resolved imported package

	Resolved bool
}

func newImport(file *File, pimp *ast.Import) (*Import, error) {
	name := pimp.Alias
	if name == "" {
		name = filepath.Base(pimp.ID)
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("empty import name")
	}

	imp := &Import{
		File: file,
		ID:   pimp.ID,
		Name: name,
	}
	return imp, nil
}

func (imp *Import) LookupType(name string) (*Definition, bool) {
	if !imp.Resolved {
		panic("import not resolved")
	}

	return imp.Package.LookupType(name)
}

func (imp *Import) Resolve(pkg *Package) error {
	if imp.Resolved {
		return fmt.Errorf("import already resolved: %v", imp.ID)
	}

	imp.Package = pkg
	imp.Resolved = true
	return nil
}
