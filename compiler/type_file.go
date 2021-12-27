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
	ImportMap map[string]*Import

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
		DefinitionNames: make(map[string]*Definition),
	}

	// create imports
	for _, pimp := range pfile.Imports {
		imp, err := newImport(f, pimp)
		if err != nil {
			return nil, fmt.Errorf("invalid file import, file=%v: %w", path, err)
		}

		_, ok := f.ImportMap[imp.Name]
		if ok {
			return nil, fmt.Errorf("duplicate file import, file=%v", path)
		}

		f.Imports = append(f.Imports, imp)
		f.ImportMap[imp.Name] = imp
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

// Import

type Import struct {
	File *File

	ID      string
	Name    string
	Package *Package // resolved imported package
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
