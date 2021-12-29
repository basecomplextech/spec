package generator

import (
	"fmt"

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
		w.line()
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
	w.linef("type %v struct {", def.Name)

	// fields
	for _, field := range def.Message.Fields {
		name := goMessageFieldName(field)
		type_ := goTypeName(field.Type)
		tag := goMessageFieldTag(field)
		w.linef("%v %v %v", name, type_, tag)
	}
	w.line("}")
	w.line()

	// write
	if err := w.messageWrite(def); err != nil {
		return err
	}
	return nil
}

func (w *goWriter) messageWrite(def *compiler.Definition) error {
	w.linef(`func (m *%v) Write(w spec.Writer) error {`, def.Name)
	w.line(`if err := w.BeginMessage(); err != nil {
		return err
	}`)
	w.line()

	for _, field := range def.Message.Fields {
		if err := w.messageWriteField(field); err != nil {
			return err
		}
	}

	w.line(`return w.EndMessage()`)
	w.line("}")
	return nil
}

func (w *goWriter) messageWriteField(field *compiler.MessageField) error {
	name := goMessageFieldName(field)
	t := field.Type
	val := fmt.Sprintf("m.%v", name)

	switch {
	case t.Bool():
		w.writeValue(t, val)
		w.linef(`w.Field(%d)`, field.Tag)

	case t.Number():
		w.linef(`if %v != 0 {`, val)
		w.writeValue(t, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case t.Bytes():
		w.linef(`if len(%v) > 0 {`, val)
		w.writeValue(t, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case t.String():
		w.linef(`if len(%v) > 0 {`, val)
		w.writeValue(t, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case t.Nullable():
		elem := t.Element
		elemVal := "(*" + val + ")"

		w.linef(`if %v != nil {`, val)
		w.writeValue(elem, elemVal)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case t.List():
		elem := t.Element

		// begin
		w.linef(`if len(%v) > 0 {`, val)
		w.line(`if err := w.BeginList(); err != nil {
			return err 
		}`)

		// elements
		w.linef(`for _, v := range %v {`, val)
		w.writeValue(elem, `v`)
		w.line(`w.Element()`)
		w.line(`}`)

		// end
		w.line(`if err := w.EndList(); err != nil {
			return err
		}`)

		// field
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case t.Enum():
		w.linef(`if %v != 0 {`, val)
		w.writeValue(t, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case t.Message():
		w.writeValue(t, val)
		w.linef(`w.Field(%d)`, field.Tag)

	case t.Struct():
		w.writeValue(t, val)
		w.linef(`w.Field(%d)`, field.Tag)
	}

	w.line()
	return nil
}

func (w *goWriter) writeValue(t *compiler.Type, val string) error {
	switch t.Kind {
	case compiler.KindBool:
		w.linef(`w.Bool(%v)`, val)
	case compiler.KindInt8:
		w.linef(`w.Int8(%v)`, val)
	case compiler.KindInt16:
		w.linef(`w.Int16(%v)`, val)
	case compiler.KindInt32:
		w.linef(`w.Int32(%v)`, val)
	case compiler.KindInt64:
		w.linef(`w.Int64(%v)`, val)

	case compiler.KindUint8:
		w.linef(`w.Uint8(%v)`, val)
	case compiler.KindUint16:
		w.linef(`w.Uint16(%v)`, val)
	case compiler.KindUint32:
		w.linef(`w.Uint32(%v)`, val)
	case compiler.KindUint64:
		w.linef(`w.Uint64(%v)`, val)

	case compiler.KindFloat32:
		w.linef(`w.Float32(%v)`, val)
	case compiler.KindFloat64:
		w.linef(`w.Float64(%v)`, val)

	case compiler.KindBytes:
		w.linef(`w.Bytes(%v)`, val)
	case compiler.KindString:
		w.linef(`w.String(%v)`, val)

	case compiler.KindReference, compiler.KindImport:
		switch {
		case t.Enum():
			w.linef(`w.Int32(int32(%v))`, val)
		case t.Message():
			w.linef(`if err := %v.Write(w); err != nil { return err }`, val)
		case t.Struct():
			w.linef(`if err := %v.Write(w); err != nil { return err }`, val)
		}

	case compiler.KindList:
		panic("cannot write list as value, write elements instead")

	case compiler.KindNullable:
		panic("cannot write nullable as value, dereference first")

	default:
		panic(fmt.Sprintf("unsupported type %v", t.Kind))
	}
	return nil
}

// struct

func (w *goWriter) struct_(def *compiler.Definition) error {
	return nil
}
