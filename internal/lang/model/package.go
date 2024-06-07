package model

import (
	"fmt"
	"path/filepath"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type PackageState string

const (
	PackageCompiling PackageState = "compiling"
	PackageCompiled  PackageState = "compiled"
)

type Package struct {
	ID    string // id is an import path as "my/example/test"
	Name  string // name is "test" in "my/example/test"
	Path  string // path is an absolute package path
	State PackageState

	Files     []*File
	FileNames map[string]*File

	Options     []*Option
	OptionNames map[string]*Option

	Definitions     []*Definition
	DefinitionNames map[string]*Definition
}

func NewPackage(id string, path string, pfiles []*syntax.File) (*Package, error) {
	name := filepath.Base(id)
	if name == "" || name == "." {
		return nil, fmt.Errorf("empty package name, id=%v, path=%v", id, path)
	}

	pkg := &Package{
		ID:    id,
		Name:  name,
		Path:  path,
		State: PackageCompiling,

		FileNames:       make(map[string]*File),
		OptionNames:     make(map[string]*Option),
		DefinitionNames: make(map[string]*Definition),
	}

	// Create files
	for _, pfile := range pfiles {
		f, err := newFile(pkg, pfile)
		if err != nil {
			return nil, err
		}

		pkg.Files = append(pkg.Files, f)
		pkg.FileNames[f.Name] = f
	}

	// Compile options
	for _, file := range pkg.Files {
		for _, opt := range file.Options {
			_, ok := pkg.OptionNames[opt.Name]
			if ok {
				return nil, fmt.Errorf("duplicate option, name=%v, path=%v", opt.Name, path)
			}

			pkg.Options = append(pkg.Options, opt)
			pkg.OptionNames[opt.Name] = opt
		}
	}

	// Compile definitions
	for _, file := range pkg.Files {
		for _, def := range file.Definitions {
			_, ok := pkg.DefinitionNames[def.Name]
			if ok {
				return nil, fmt.Errorf("duplicate definition, name=%v, path=%v", def.Name, path)
			}

			pkg.Definitions = append(pkg.Definitions, def)
			pkg.DefinitionNames[def.Name] = def
		}
	}
	return pkg, nil
}

func (p *Package) ResolveImports(getPackage func(string) (*Package, error)) error {
	for _, file := range p.Files {
		if err := file.resolveImports(getPackage); err != nil {
			return fmt.Errorf("%v: %w", p.Name, err)
		}
	}
	return nil
}

func (p *Package) ResolveTypes() error {
	for _, file := range p.Files {
		if err := file.resolve(); err != nil {
			return fmt.Errorf("%v: %w", p.Name, err)
		}
	}
	return nil
}

func (p *Package) Resolved() error {
	for _, file := range p.Files {
		if err := file.resolved(); err != nil {
			return fmt.Errorf("%v: %w", p.Name, err)
		}
	}
	return nil
}

func (p *Package) LookupType(name string) (*Definition, bool) {
	def, ok := p.DefinitionNames[name]
	return def, ok
}

// internal

func (p *Package) getType(name string) (*Definition, error) {
	def, ok := p.DefinitionNames[name]
	if !ok {
		return nil, fmt.Errorf("type not found: %v", name)
	}
	return def, nil
}
