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

	if err := c._resolveImports(pkg); err != nil {
		return nil, err
	}
	if err := c._resolveTypes(pkg); err != nil {
		return nil, err
	}
	if err := c._resolved(pkg); err != nil {
		return nil, err
	}

	// Done
	pkg.State = model.PackageCompiled
	return pkg, nil
}

func (c *compiler) _resolveImports(pkg *model.Package) error {
	for _, file := range pkg.Files {
		for _, imp := range file.Imports {
			if err := c._resolveImport(imp); err != nil {
				return fmt.Errorf("%v/%v: %w", pkg.Name, file.Name, err)
			}
		}
	}
	return nil
}

func (c *compiler) _resolveImport(imp *model.Import) error {
	id := imp.ID

	pkg, err := c.getPackage(id)
	if err != nil {
		return err
	}

	return imp.Resolve(pkg)
}

func (c *compiler) _resolveTypes(pkg *model.Package) error {
	for _, file := range pkg.Files {
		for _, def := range file.Definitions {
			if err := c._resolveDefinition(file, def); err != nil {
				return fmt.Errorf("%v/%v: %w", pkg.Name, file.Name, err)
			}
		}
	}
	return nil
}

func (c *compiler) _resolveDefinition(file *model.File, def *model.Definition) error {
	switch def.Type {
	case model.DefinitionMessage:
		return c._resolveMessage(file, def)
	case model.DefinitionStruct:
		return c._resolveStruct(file, def)
	case model.DefinitionService:
		return c._resolveService(file, def)
	}
	return nil
}

func (c *compiler) _resolveMessage(file *model.File, def *model.Definition) error {
	for _, field := range def.Message.Fields {
		if err := c._resolveType(file, field.Type); err != nil {
			return fmt.Errorf("%v.%v: %w", def.Name, field.Name, err)
		}
	}
	return nil
}

func (c *compiler) _resolveStruct(file *model.File, def *model.Definition) error {
	for _, field := range def.Struct.Fields {
		if err := c._resolveType(file, field.Type); err != nil {
			return fmt.Errorf("%v.%v: %w", def.Name, field.Name, err)
		}
	}
	return nil
}

func (c *compiler) _resolveService(file *model.File, def *model.Definition) error {
	for _, method := range def.Service.Methods {
		for _, arg := range method.Args {
			if err := c._resolveType(file, arg.Type); err != nil {
				return fmt.Errorf("%v.%v %v: %w", def.Name, method.Name, arg.Name, err)
			}
		}
		for _, result := range method.Results {
			if err := c._resolveType(file, result.Type); err != nil {
				return fmt.Errorf("%v.%v %v: %w", def.Name, method.Name, result.Name, err)
			}
		}
	}
	return nil
}

func (c *compiler) _resolveType(file *model.File, type_ *model.Type) error {
	switch type_.Kind {
	case model.KindList:
		return c._resolveType(file, type_.Element)

	case model.KindReference:
		if type_.ImportName == "" {
			// Local type

			pkg := file.Package
			def, ok := pkg.LookupType(type_.Name)
			if !ok {
				return fmt.Errorf("type not found: %v", type_.Name)
			}
			type_.Resolve(def, nil)

		} else {
			// Imported type

			imp, ok := file.LookupImport(type_.ImportName)
			if !ok {
				return fmt.Errorf("type not found: %v.%v", type_.ImportName, type_.Name)
			}
			def, ok := imp.LookupType(type_.Name)
			if !ok {
				return fmt.Errorf("type not found: %v.%v", type_.ImportName, type_.Name)
			}
			type_.Resolve(def, imp)
		}
	}
	return nil
}

// resolved

func (c *compiler) _resolved(pkg *model.Package) error {
	for _, file := range pkg.Files {
		for _, def := range file.Definitions {
			switch def.Type {
			case model.DefinitionService:
				if err := def.Service.Resolved(); err != nil {
					return fmt.Errorf("%v/%v: %w", pkg.Name, file.Name, err)
				}
			}
		}
	}
	return nil
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
