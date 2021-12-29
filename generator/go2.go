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

// file

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

// enum

func (w *goWriter) enum(def *compiler.Definition) error {
	w.linef("type %v int32", def.Name)
	w.line()

	// values
	w.line("const (")
	for _, val := range def.Enum.Values {
		// EnumValue Enum = 1
		name := goEnumValueName(val)
		w.linef("%v %v = %d", name, def.Name, val.Number)
	}
	w.line(")")
	w.line()

	// string
	w.linef("func (e %v) String() string {", def.Name)
	w.line("switch e {")
	for _, val := range def.Enum.Values {
		name := goEnumValueName(val)
		w.linef("case %v:", name)
		w.linef(`return "%v"`, toLowerCase(val.Name))
	}
	w.line("}")
	w.line(`return ""`)
	w.line("}")
	return nil
}

// message

func (w *goWriter) message(def *compiler.Definition) error {
	return nil
}

// struct

func (w *goWriter) struct_(def *compiler.Definition) error {
	return nil
}
