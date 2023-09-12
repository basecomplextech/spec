package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"

	"github.com/basecomplextech/spec/internal/lang/model"
)

const (
	OptionPackage = "go_package"
)

// WriteFile writes a golang file.
func WriteFile(file *model.File, skipRPC bool) ([]byte, error) {
	w := newWriter(skipRPC)
	if err := w.file(file); err != nil {
		return nil, err
	}

	bytes := w.b.Bytes()
	return format.Source(bytes)
}

type writer struct {
	b bytes.Buffer

	skipRPC bool
}

func newWriter(skipRPC bool) *writer {
	return &writer{
		b: bytes.Buffer{},

		skipRPC: skipRPC,
	}
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

func (w *writer) file(file *model.File) error {
	// Package
	w.line("package ", file.Package.Name)
	w.line()

	// Imports
	w.line("import (")
	w.line(`"github.com/basecomplextech/baselibrary/bin"`)
	w.line(`"github.com/basecomplextech/baselibrary/buffer"`)
	w.line(`"github.com/basecomplextech/baselibrary/status"`)
	w.line(`"github.com/basecomplextech/spec"`)
	w.line(`"github.com/basecomplextech/spec/encoding"`)

	if !w.skipRPC {
		w.line(`"github.com/basecomplextech/spec/rpc"`)
	}

	for _, imp := range file.Imports {
		pkg := importPackage(imp)
		w.linef(`"%v"`, pkg)
	}
	w.line(")")
	w.line()

	// Empty values for imports
	w.line(`var (`)
	w.line(`_ bin.Bin128`)
	w.line(`_ buffer.Buffer`)
	w.line(`_ encoding.Type`)

	if !w.skipRPC {
		w.line(`_ rpc.Client`)
	}

	w.line(`_ spec.Type`)
	w.line(`_ status.Status`)
	w.line(`)`)

	// Definitions
	return w.definitions(file)
}

func (w *writer) definitions(file *model.File) error {
	for _, def := range file.Definitions {
		switch def.Type {
		case model.DefinitionEnum:
			if err := w.enum(def); err != nil {
				return err
			}
		case model.DefinitionMessage:
			if err := w.message(def); err != nil {
				return err
			}
		case model.DefinitionStruct:
			if err := w.struct_(def); err != nil {
				return err
			}
		case model.DefinitionService:
			if !w.skipRPC {
				// TODO: Uncomment
				// if err := w.service(def); err != nil {
				// 	return err
				// }
			}
		}
	}

	for _, def := range file.Definitions {
		switch def.Type {
		case model.DefinitionMessage:
			if err := w.messageWriter(def); err != nil {
				return err
			}
		}
	}
	return nil
}

// internal

func importPackage(imp *model.Import) string {
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

func toLowerCameCase(s string) string {
	if len(s) == 0 {
		return ""
	}

	s = toUpperCamelCase(s)
	return strings.ToLower(s[:1]) + s[1:]
}
