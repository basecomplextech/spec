// Copyright 2023 Ivan Korobkov. All rights reserved.

package model

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type Import struct {
	File *File

	ID      string   // full id
	Name    string   // name or alias
	Package *Package // resolved imported package

	Resolved bool
}

func newImport(file *File, pimp *syntax.Import) (*Import, error) {
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

func (imp *Import) lookupType(name string) (*Definition, bool) {
	if !imp.Resolved {
		panic("import not resolved")
	}

	return imp.Package.lookupType(name)
}

// internal

func (imp *Import) resolve() error {
	if imp.Resolved {
		return fmt.Errorf("import already resolved: %v", imp.ID)
	}

	ctx := imp.File.Package.Context
	pkg, err := ctx.getPackage(imp.ID)
	if err != nil {
		return err
	}

	imp.Package = pkg
	imp.Resolved = true
	return nil
}
