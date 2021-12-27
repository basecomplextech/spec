package compiler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/baseone-run/spec/parser"
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

	Files []*File

	Definitions   []*Definition
	DefinitionMap map[string]*Definition
}

func newPackage(id string, path string, pfiles []*parser.File) (*Package, error) {
	name := filepath.Base(id)
	if name == "" || name == "." {
		return nil, fmt.Errorf("invalid package name, id=%v, path=%v", id, path)
	}

	pkg := &Package{
		ID:    id,
		Name:  name,
		Path:  path,
		State: PackageCompiling,

		DefinitionMap: make(map[string]*Definition),
	}

	// create files
	for _, pfile := range pfiles {
		f, err := newFile(pkg, pfile)
		if err != nil {
			return nil, err
		}
		pkg.Files = append(pkg.Files, f)
	}

	// compile definitions
	for _, file := range pkg.Files {
		for _, def := range file.Definitions {

			_, ok := pkg.DefinitionMap[def.Name]
			if ok {
				return nil, fmt.Errorf("duplicate definition in package, name=%v, path=%v",
					def.Name, path)
			}

			pkg.Definitions = append(pkg.Definitions, def)
			pkg.DefinitionMap[def.Name] = def
		}
	}
	return pkg, nil
}

// File

type File struct {
	Package *Package
	Name    string
	Path    string

	Imports   []*Import
	ImportMap map[string]*Import

	Definitions   []*Definition
	DefinitionMap map[string]*Definition
}

func newFile(pkg *Package, pfile *parser.File) (*File, error) {
	path := pfile.Path
	name := filepath.Base(path)

	f := &File{
		Package: pkg,
		Name:    name,
		Path:    path,

		ImportMap:     make(map[string]*Import),
		DefinitionMap: make(map[string]*Definition),
	}

	// create imports
	for _, pimp := range pfile.Imports {
		imp, err := newImport(f, pimp)
		if err != nil {
			return nil, fmt.Errorf("invalid import in file, path=%v: %w", path, err)
		}

		_, ok := f.ImportMap[imp.Name]
		if ok {
			return nil, fmt.Errorf("duplicate import in file, path=%v", path)
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

		_, ok := f.DefinitionMap[def.Name]
		if ok {
			return nil, fmt.Errorf("duplicate definition in file, path=%v", path)
		}

		f.Definitions = append(f.Definitions, def)
		f.DefinitionMap[def.Name] = def
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

// Definition type

type DefinitionType string

const (
	DefinitionUndefined DefinitionType = ""
	DefinitionEnum      DefinitionType = "enum"
	DefinitionMessage   DefinitionType = "message"
	DefinitionStruct    DefinitionType = "struct"
)

func getDefinitionType(ptype parser.DefinitionType) (DefinitionType, error) {
	switch ptype {
	case parser.DefinitionEnum:
		return DefinitionEnum, nil
	case parser.DefinitionMessage:
		return DefinitionMessage, nil
	case parser.DefinitionStruct:
		return DefinitionStruct, nil
	}
	return "", fmt.Errorf("unsupported parser definition type %v", ptype)
}

// Definition

type Definition struct {
	Pkg  *Package
	File *File

	Name string
	Type DefinitionType

	Enum    *Enum
	Message *Message
	Struct  *Struct
}

func newDefinition(pkg *Package, file *File, pdef *parser.Definition) (*Definition, error) {
	type_, err := getDefinitionType(pdef.Type)
	if err != nil {
		return nil, err
	}

	def := &Definition{
		Pkg:  pkg,
		File: file,

		Name: pdef.Name,
		Type: type_,
	}

	switch type_ {
	case DefinitionEnum:
		def.Enum, err = newEnum(pdef.Enum)
		if err != nil {
			return nil, err
		}

	case DefinitionMessage:
		def.Message, err = newMessage(pdef.Message)
		if err != nil {
			return nil, err
		}

	case DefinitionStruct:
		def.Struct, err = newStruct(pdef.Struct)
		if err != nil {
			return nil, err
		}
	}

	return def, nil
}

// Enum

type Enum struct {
}

func newEnum(penum *parser.Enum) (*Enum, error) {
	return nil, nil
}

// Message

type Message struct {
}

func newMessage(pmsg *parser.Message) (*Message, error) {
	return nil, nil
}

// Struct

type Struct struct {
}

func newStruct(pstr *parser.Struct) (*Struct, error) {
	return nil, nil
}
