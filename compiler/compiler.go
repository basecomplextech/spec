package compiler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/baseone-run/spec/parser"
)

type Compiler interface {
	// Compile parses, compiles and returns a package.
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

// Compile parses, compiles and returns a package.
func (c *compiler) Compile(path string) (*Package, error) {
	// clean path
	path = filepath.Clean(path)

	// get absolute path relative to cwd
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// compute package path from directory name
	// when path is empty
	pkgPath := path
	if pkgPath == "" || pkgPath == "." {
		pkgPath = filepath.Base(absPath)
	}

	return c.compile(path, absPath)
}

// private

func (c *compiler) compile(pkgPath string, absPath string) (*Package, error) {
	// return a package if exists already
	pkg, ok := c.packages[pkgPath]
	if ok {
		return pkg, nil
	}

	// check package is dir
	info, err := os.Stat(absPath)
	switch {
	case err != nil:
		return nil, err
	case !info.IsDir():
		return nil, fmt.Errorf("compile: not directory, abs path=%v", absPath)
	}

	// scan files
	pattern := filepath.Join(absPath, "*.spec")
	filepaths, err := filepath.Glob(pattern)
	switch {
	case err != nil:
		return nil, err
	case len(filepaths) == 0:
		return nil, fmt.Errorf("compile: empty package, abs path=%v", absPath)
	}

	// create package
	pkg, err = newPackage(pkgPath, absPath)
	if err != nil {
		return nil, err
	}
	c.packages[pkgPath] = pkg

	// parse files
	for _, filepath := range filepaths {
		file, err := c.parser.ParsePath(filepath)
		if err != nil {
			return nil, err
		}

		pkg.Files = append(pkg.Files, file)
	}

	return pkg, nil
}
