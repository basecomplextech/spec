package compiler

import (
	"fmt"
	"os"
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

	packages map[string]*Package // compiled packages by ids
	paths    []string            // import paths
}

func newCompiler(opts Options) (*compiler, error) {
	parser := parser.New()

	paths := make([]string, 0, len(opts.ImportPath))
	for _, path := range opts.ImportPath {
		_, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("invalid import path %q: %w", path, err)
		}

		paths = append(paths, path)
	}

	c := &compiler{
		opts:   opts,
		parser: parser,

		packages: make(map[string]*Package),
		paths:    paths,
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

// private

func (c *compiler) getPackage(id string) (*Package, error) {
	// try to get existing package
	pkg, ok := c.packages[id]
	if ok {
		if pkg.State != PackageCompiled {
			return nil, fmt.Errorf("circular import: %v", id)
		}
		return pkg, nil
	}

	// try to find package in import paths
	for _, path := range c.paths {
		p := filepath.Join(path, id)
		_, err := os.Stat(p)
		switch {
		case os.IsNotExist(err):
			continue
		case err != nil:
			return nil, err
		}

		// found package
		return c.compilePackage(id, p)
	}

	return nil, fmt.Errorf("package not found: %v", id)
}

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
		return nil, fmt.Errorf("empty package %q, path=%v", id, path)
	}

	// create package in compiling state
	pkg, err = newPackage(id, path, files)
	if err != nil {
		return nil, fmt.Errorf("invalid package %q: %w", id, err)
	}
	c.packages[id] = pkg

	if err := c._resolveImports(pkg); err != nil {
		return nil, fmt.Errorf("invalid package %q: %w", id, err)
	}
	if err := c._resolveTypes(pkg); err != nil {
		return nil, fmt.Errorf("invalid package %q: %w", id, err)
	}

	// done
	pkg.State = PackageCompiled
	return pkg, nil
}

func (c *compiler) _resolveImports(pkg *Package) error {
	for _, file := range pkg.Files {
		for _, imp := range file.Imports {
			if err := c._resolveImport(imp); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *compiler) _resolveImport(imp *Import) error {
	id := imp.ID

	pkg, err := c.getPackage(id)
	if err != nil {
		return err
	}

	return imp.resolve(pkg)
}

func (c *compiler) _resolveTypes(pkg *Package) error {
	return nil
}
