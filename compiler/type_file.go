package compiler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/baseone-run/spec/parser"
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

func newFile(pkg *Package, pfile *parser.File) (*File, error) {
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

	// create imports
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

	// create options
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

	// create definitions
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

func (f *File) lookupType(name string) (*Definition, bool) {
	def, ok := f.DefinitionNames[name]
	return def, ok
}

func (f *File) lookupImport(name string) (*Import, bool) {
	imp, ok := f.ImportMap[name]
	return imp, ok
}

// Import

type Import struct {
	File *File

	ID      string   // full id
	Name    string   // name or alias
	Package *Package // resolved imported package

	Resolved bool
}

func newImport(file *File, pimp *parser.Import) (*Import, error) {
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

func (imp *Import) resolve(pkg *Package) error {
	if imp.Resolved {
		return fmt.Errorf("import already resolved: %v", imp.ID)
	}

	imp.Package = pkg
	imp.Resolved = true
	return nil
}

func (imp *Import) lookupType(name string) (*Definition, bool) {
	if !imp.Resolved {
		panic("import not resolved")
	}

	return imp.Package.lookupType(name)
}
