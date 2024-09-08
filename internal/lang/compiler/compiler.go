// Copyright 2021 Ivan Korobkov. All rights reserved.

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
	paths  []string // import paths
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
		paths:  paths,
	}
	return c, nil
}

// Compile parses, compiles and returns a package from a directory.
func (c *compiler) Compile(dir string) (*model.Package, error) {
	// Clean directory path
	dir = filepath.Clean(dir)

	// Compute id from directory when empty
	id := dir
	var err error
	if id == "" || id == "." {
		id, err = getCurrentDirectoryName()
		if err != nil {
			return nil, err
		}
	}

	x := model.NewContext(c.parser, c.paths)
	return x.Compile(id, dir)
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
