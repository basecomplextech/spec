package model

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/basecomplextech/spec/internal/lang/parser"
	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type Context struct {
	Parser      parser.Parser
	ImportPaths []string // import paths

	Packages map[string]*Package // compiled packages by ids
}

// NewContext returns a new package context.
func NewContext(parser parser.Parser, importPaths []string) *Context {
	return &Context{
		Parser:      parser,
		ImportPaths: importPaths,

		Packages: make(map[string]*Package),
	}
}

// Compile compiles a package from a directory.
func (x *Context) Compile(id string, path string) (*Package, error) {
	return x.compile(id, path)
}

// internal

func (x *Context) getPackage(id string) (*Package, error) {
	// Maybe already exists
	pkg, ok := x.Packages[id]
	if ok {
		if pkg.Compiling {
			return nil, fmt.Errorf("circular import: %v", id)
		}
		return pkg, nil
	}

	// Try to find package in import paths
	for _, path := range x.ImportPaths {
		p := filepath.Join(path, id)
		_, err := os.Stat(p)
		switch {
		case os.IsNotExist(err):
			continue
		case err != nil:
			return nil, err
		}

		return x.compile(id, p)
	}

	return nil, fmt.Errorf("package not found: %v", id)
}

// compile

func (x *Context) compile(id string, path string) (*Package, error) {
	// Return if already exists
	pkg, ok := x.Packages[id]
	if ok {
		return pkg, nil
	}

	// Parse directory files
	files, err := x.Parser.ParseDirectory(path)
	if err != nil {
		return nil, err
	}
	return x.compileFiles(id, path, files)
}

func (x *Context) compileFiles(id string, path string, files []*syntax.File) (*Package, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("%v: empty package", path)
	}

	// Check not exists
	_, ok := x.Packages[id]
	if ok {
		return nil, fmt.Errorf("%v: duplicate package %q", path, id)
	}

	// Parse package
	pkg, err := parsePackage(x, id, path, files)
	if err != nil {
		return nil, err
	}
	x.Packages[id] = pkg

	// Resolve, compile, validate
	if err := pkg.resolve(); err != nil {
		return nil, err
	}
	if err := pkg.compile(); err != nil {
		return nil, err
	}
	if err := pkg.validate(); err != nil {
		return nil, err
	}

	// Done
	pkg.Compiling = false
	return pkg, nil
}
