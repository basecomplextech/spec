package compiler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/basecomplextech/spec/internal/lang/model"
	"github.com/basecomplextech/spec/internal/lang/parser"
)

type Compiler interface {
	// Compile parses, compiles and returns a package from a directory.
	Compile(path string) (*model.Package, error)
}

type Options struct {
	ImportPath []string
}

// New returns a new compiler.
func New(opts Options) (Compiler, error) {
	return newCompiler(opts)
}

type compiler struct {
	opts   Options
	parser parser.Parser

	packages map[string]*model.Package // compiled packages by ids
	paths    []string                  // import paths
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

		packages: make(map[string]*model.Package),
		paths:    paths,
	}
	return c, nil
}

// Compile parses, compiles and returns a package from a directory.
func (c *compiler) Compile(dir string) (*model.Package, error) {
	// Clean directory path
	dir = filepath.Clean(dir)

	// Get absolute path relative to cwd
	path, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	// Compute id from directory when empty
	id := dir
	if id == "" || id == "." {
		id, err = getCurrentDirectoryName()
		if err != nil {
			return nil, err
		}
	}

	return c.compilePackage(id, path)
}

// private

func (c *compiler) getPackage(id string) (*model.Package, error) {
	// Try to get existing package
	pkg, ok := c.packages[id]
	if ok {
		if pkg.State != model.PackageCompiled {
			return nil, fmt.Errorf("circular import: %v", id)
		}
		return pkg, nil
	}

	// Try to find package in import paths
	for _, path := range c.paths {
		p := filepath.Join(path, id)
		_, err := os.Stat(p)
		switch {
		case os.IsNotExist(err):
			continue
		case err != nil:
			return nil, err
		}

		// Found package
		return c.compilePackage(id, p)
	}

	return nil, fmt.Errorf("package not found: %v", id)
}

func (c *compiler) compilePackage(id string, path string) (*model.Package, error) {
	// Return if already exists
	pkg, ok := c.packages[id]
	if ok {
		return pkg, nil
	}

	// Parse directory files
	files, err := c.parser.ParseDirectory(path)
	switch {
	case err != nil:
		return nil, err
	case len(files) == 0:
		return nil, fmt.Errorf("empty package %q, path=%v", id, path)
	}

	// Create package in compiling state
	pkg, err = model.NewPackage(id, path, files)
	if err != nil {
		return nil, err
	}
	c.packages[id] = pkg

	if err := pkg.ResolveImports(c.getPackage); err != nil {
		return nil, err
	}
	if err := pkg.ResolveTypes(); err != nil {
		return nil, err
	}
	if err := pkg.Resolved(); err != nil {
		return nil, err
	}

	// Done
	pkg.State = model.PackageCompiled
	return pkg, nil
}

// private

func getCurrentDirectoryName() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	dir = filepath.Base(dir)
	return dir, nil
}
