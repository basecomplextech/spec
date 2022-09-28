package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"

	"github.com/complex1tech/spec/lang/compiler"
)

const (
	OptionPackage = "go_package"
)

// WriteFile writes a golang file.
func WriteFile(file *compiler.File) ([]byte, error) {
	w := newWriter()
	if err := w.file(file); err != nil {
		return nil, err
	}

	bytes := w.b.Bytes()
	return format.Source(bytes)
}

type writer struct {
	b bytes.Buffer
}

func newWriter() *writer {
	return &writer{b: bytes.Buffer{}}
}

func (w *writer) line(args ...string) {
	w.write(args...)
	w.b.WriteString("\n")
}

func (w *writer) linef(format string, args ...interface{}) {
	w.writef(format, args...)
	w.b.WriteString("\n")
}

func (w *writer) write(args ...string) {
	for _, s := range args {
		w.b.WriteString(s)
	}
}

func (w *writer) writef(format string, args ...interface{}) {
	if len(args) == 0 {
		w.write(format)
		return
	}

	s := fmt.Sprintf(format, args...)
	w.b.WriteString(s)
}

// file

func (w *writer) file(file *compiler.File) error {
	// package
	w.line("package ", file.Package.Name)
	w.line()

	// imports
	w.line("import (")
	w.line(`"github.com/complex1tech/baselibrary/buffer"`)
	w.line(`"github.com/complex1tech/baselibrary/types"`)
	w.line(`spec "github.com/complex1tech/spec/go"`)

	for _, imp := range file.Imports {
		pkg := importPackage(imp)
		w.linef(`"%v"`, pkg)
	}
	w.line(")")
	w.line()

	// empty values for imports
	w.line(`var (`)
	w.line(`_ types.Bin128 = types.Bin128{}`)
	w.line(`_ buffer.Buffer = (buffer.Buffer)(nil)`)
	w.line(`)`)

	// definitions
	return w.definitions(file)
}

func (w *writer) definitions(file *compiler.File) error {
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

	for _, def := range file.Definitions {
		switch def.Type {
		case compiler.DefinitionMessage:
			if err := w.messageBuilder(def); err != nil {
				return err
			}
		}
	}
	return nil
}

// internal

func importPackage(imp *compiler.Import) string {
	pkg, ok := imp.Package.OptionNames[OptionPackage]
	if ok {
		return pkg.Value
	}

	return imp.ID
}

func toUpperCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		part = strings.ToLower(part)
		part = strings.Title(part)
		parts[i] = part
	}
	return strings.Join(parts, "")
}
