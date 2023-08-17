package compiler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/basecomplextech/spec/lang/parser"
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
	return newCompiler(opts)
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

func (c *compiler) getPackage(id string) (*Package, error) {
	// Try to get existing package
	pkg, ok := c.packages[id]
	if ok {
		if pkg.State != PackageCompiled {
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

func (c *compiler) compilePackage(id string, path string) (*Package, error) {
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
	if err := c._resolved(pkg); err != nil {
		return nil, err
	}

	// Done
	pkg.State = PackageCompiled
	return pkg, nil
}

func (c *compiler) _resolveImports(pkg *Package) error {
	for _, file := range pkg.Files {
		for _, imp := range file.Imports {
			if err := c._resolveImport(imp); err != nil {
				return fmt.Errorf("%v/%v: %w", pkg.Name, file.Name, err)
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
	for _, file := range pkg.Files {
		for _, def := range file.Definitions {
			if err := c._resolveDefinition(file, def); err != nil {
				return fmt.Errorf("%v/%v: %w", pkg.Name, file.Name, err)
			}
		}
	}
	return nil
}

func (c *compiler) _resolveDefinition(file *File, def *Definition) error {
	switch def.Type {
	case DefinitionMessage:
		return c._resolveMessage(file, def)
	case DefinitionStruct:
		return c._resolveStruct(file, def)
	case DefinitionService:
		return c._resolveService(file, def)
	}
	return nil
}

func (c *compiler) _resolveMessage(file *File, def *Definition) error {
	for _, field := range def.Message.Fields {
		if err := c._resolveType(file, field.Type); err != nil {
			return fmt.Errorf("%v.%v: %w", def.Name, field.Name, err)
		}
	}
	return nil
}

func (c *compiler) _resolveStruct(file *File, def *Definition) error {
	for _, field := range def.Struct.Fields {
		if err := c._resolveType(file, field.Type); err != nil {
			return fmt.Errorf("%v.%v: %w", def.Name, field.Name, err)
		}
	}
	return nil
}

func (c *compiler) _resolveService(file *File, def *Definition) error {
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

func (c *compiler) _resolveType(file *File, type_ *Type) error {
	switch type_.Kind {
	case KindList:
		return c._resolveType(file, type_.Element)

	case KindReference:
		if type_.ImportName == "" {
			// Local type

			pkg := file.Package
			def, ok := pkg.lookupType(type_.Name)
			if !ok {
				return fmt.Errorf("type not found: %v", type_.Name)
			}
			type_.resolve(def, nil)

		} else {
			// Imported type

			imp, ok := file.lookupImport(type_.ImportName)
			if !ok {
				return fmt.Errorf("type not found: %v.%v", type_.ImportName, type_.Name)
			}
			def, ok := imp.lookupType(type_.Name)
			if !ok {
				return fmt.Errorf("type not found: %v.%v", type_.ImportName, type_.Name)
			}
			type_.resolve(def, imp)
		}
	}
	return nil
}

// resolved

func (c *compiler) _resolved(pkg *Package) error {
	for _, file := range pkg.Files {
		for _, def := range file.Definitions {
			switch def.Type {
			case DefinitionService:
				if err := def.Service.resolved(); err != nil {
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
