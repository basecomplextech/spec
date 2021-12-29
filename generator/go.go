package generator

import (
	"github.com/baseone-run/spec/compiler"
)

// GenerateGo generates a go package.
func (g *generator) GenerateGo(pkg *compiler.Package) error {
	for _, file := range pkg.Files {
		if err := g.generateGoFile(file); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) generateGoFile(file *compiler.File) error {
	w := newGoWriter()
	if err := w.file(file); err != nil {
		return err
	}

	path := filenameWithExt(file.Name, "go")
	f, err := g.createFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = w.b.WriteTo(f)
	return err
}

type goWriter struct {
	*writer
}

func newGoWriter() *goWriter {
	w := newWriter()
	return &goWriter{writer: w}
}

func (w *goWriter) file(file *compiler.File) error {
	// package
	w.line("package ", file.Package.Name)
	w.line()

	// imports
	w.line("import (")
	w.line(`"github.com/baseone-run/spec"`)
	for _, imp := range file.Imports {
		w.linef(`"%v"`, imp.ID)
	}
	w.line(")")
	w.line()

	// definitions
	return w.definitions(file)
}

func (w *goWriter) definitions(file *compiler.File) error {
	for _, def := range file.Definitions {
		switch def.Type {
		case compiler.DefinitionEnum:
			if err := w.enum(def); err != nil {
				return err
			}
		case compiler.DefinitionMessage:
			if err := w.message(def); err != nil {
				return err
			}
		case compiler.DefinitionStruct:
			if err := w.struct_(def); err != nil {
				return err
			}
		}
	}
	return nil
}
