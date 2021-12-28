package compiler

import (
	"fmt"
	"path/filepath"

	"github.com/baseone-run/spec/parser"
)

type Compiler interface {
	// Compile parses, compiles and returns a package from a directory.
	Compile(path string) (*Package, error)
}

type Options struct {
	ImportPath []string
}

// New returns a new compiler.
func New(opts Options) (Compiler, error) {
	return nil, nil
}

type compiler struct {
	opts   Options
	parser parser.Parser

	packages map[string]*Package
}

func newCompiler(opts Options) (*compiler, error) {
	parser := parser.New()

	c := &compiler{
		opts:   opts,
		parser: parser,

		packages: make(map[string]*Package),
	}
	return c, nil
}

// Compile parses, compiles and returns a package from a directory.
func (c *compiler) Compile(dir string) (*Package, error) {
	// clean directory path
	dir = filepath.Clean(dir)

	// get absolute path relative to cwd
	path, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	// compute id from directory
	// when current dir
	id := dir
	if id == "" || id == "." {
		id = filepath.Base(dir)
	}

	return c.compilePackage(id, path)
}

// package

func (c *compiler) compilePackage(id string, path string) (*Package, error) {
	// return if already exists
	pkg, ok := c.packages[id]
	if ok {
		return pkg, nil
	}

	// parse directory files
	files, err := c.parser.ParseDirectory(path)
	switch {
	case err != nil:
		return nil, err
	case len(files) == 0:
		return nil, fmt.Errorf("package is empty, path=%v", path)
	}

	// create package in compiling state
	pkg, err = newPackage(id, path, files)
	if err != nil {
		return nil, err
	}
	c.packages[id] = pkg

	if err := c._resolveImports(pkg); err != nil {
		return nil, err
	}
	if err := c._resolveTypes(pkg); err != nil {
		return nil, err
	}

	// done
	pkg.State = PackageCompiled
	return pkg, nil
}

func (c *compiler) _resolveImports(pkg *Package) error {
	return nil
}

func (c *compiler) _resolveTypes(pkg *Package) error {
	return nil
}

// import

func (c *compiler) resolveImport(imp *Import) (*Package, error) {
	return nil, nil
}
