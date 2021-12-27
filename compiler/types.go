package compiler

import (
	"fmt"
	"path/filepath"

	"github.com/baseone-run/spec/parser"
)

type Package struct {
	Name    string // name is a directory name as "test" in "my/example/test"
	Path    string // path is an import path as "my/example/test"
	AbsPath string // abs path is an absolute directory path

	Files []*parser.File
}

func newPackage(path string, absPath string) (*Package, error) {
	name := filepath.Base(path)
	if name == "" || name == "." {
		return nil, fmt.Errorf("invalid package name, path=%v, abs path=%v", path, absPath)
	}

	pkg := &Package{
		Name:    name,
		Path:    path,
		AbsPath: absPath,
	}
	return pkg, nil
}

type Module struct {
	Name    string
	Imports []*Import

	Definitionis []*Definition
}

type Import struct {
}

type DefinitionType string

const (
	DefinitionUndefined DefinitionType = ""
	DefinitionEnum      DefinitionType = "enum"
	DefinitionMessage   DefinitionType = "message"
	DefinitionStruct    DefinitionType = "struct"
)

type Definition struct {
	Type DefinitionType

	Enum    *Enum
	Message *Message
	Struct  *Struct
}

type Message struct {
}

type Enum struct {
}

type Struct struct {
}
