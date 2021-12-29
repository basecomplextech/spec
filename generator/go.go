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
	if err := w.definitions(file); err != nil {
		return err
	}

	// reads
	if err := w.reads(file); err != nil {
		return err
	}

	// writes
	if err := w.writes(file); err != nil {
		return err
	}

	return nil
}

// definitions

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
	w.line()
	return nil
}

// message

func (w *goWriter) message(def *compiler.Definition) error {
	w.linef("type %v struct {", def.Name)

	for _, field := range def.Message.Fields {
		name := goMessageFieldName(field)
		type_ := goTypeName(field.Type)
		tag := fmt.Sprintf("`tag:\"%d\" json:\"%v\"`", field.Tag, field.Name)
		w.linef("%v %v %v", name, type_, tag)
	}

	w.line("}")
	w.line()
	return nil
}

// reads

func (w *goWriter) reads(file *compiler.File) error {
	w.line("// Read")
	w.line()

	for _, def := range file.Definitions {
		switch def.Type {
		case compiler.DefinitionEnum:
			// enum has no read method

		case compiler.DefinitionMessage:
			if err := w.readMessage(def); err != nil {
				return err
			}
			if err := w.readMessageValue(def); err != nil {
				return err
			}
			if err := w.readMessageMethod(def); err != nil {
				return err
			}
		case compiler.DefinitionStruct:
			if err := w.readStruct(def); err != nil {
				return err
			}
		}
	}
	return nil
}

// read message

func (w *goWriter) readMessage(def *compiler.Definition) error {
	w.linef(`func Read%v(b []byte) (*%v, error) {`, def.Name, def.Name)
	w.linef(`if len(b) == 0 {
		return nil, nil
	}`)
	w.line()
	w.linef(`m := &%v{}`, def.Name)
	w.line(`if err := m.Read(b); err != nil {
		return nil, err
	}`)
	w.line(`return m, nil`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *goWriter) readMessageValue(def *compiler.Definition) error {
	w.linef(`func Read%vValue(b []byte) (m %v, err error) {`, def.Name, def.Name)
	w.linef(`if len(b) == 0 {
		return
	}`)
	w.line()
	w.line(`err = m.Read(b)`)
	w.line(`return`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *goWriter) readMessageMethod(def *compiler.Definition) error {
	w.linef(`func (m *%v) Read(b []byte) error {`, def.Name)
	w.line(`msg, err := spec.ReadMessage(b)
	if err != nil {
		return err
	}`)
	w.line()

	for _, field := range def.Message.Fields {
		if err := w.readMessageField(field); err != nil {
			return err
		}
	}

	w.line(`return nil`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *goWriter) readMessageField(field *compiler.MessageField) error {
	name := goMessageFieldName(field)
	kind := field.Type.Kind
	typ := field.Type
	tag := field.Tag

	switch kind {
	case compiler.KindBool:
		w.linef(`m.%v = msg.Bool(%d)`, name, tag)

	case compiler.KindInt8:
		w.linef(`m.%v = msg.Int8(%d)`, name, tag)
	case compiler.KindInt16:
		w.linef(`m.%v = msg.Int16(%d)`, name, tag)
	case compiler.KindInt32:
		w.linef(`m.%v = msg.Int32(%d)`, name, tag)
	case compiler.KindInt64:
		w.linef(`m.%v = msg.Int64(%d)`, name, tag)

	case compiler.KindUint8:
		w.linef(`m.%v = msg.Uint8(%d)`, name, tag)
	case compiler.KindUint16:
		w.linef(`m.%v = msg.Uint16(%d)`, name, tag)
	case compiler.KindUint32:
		w.linef(`m.%v = msg.Uint32(%d)`, name, tag)
	case compiler.KindUint64:
		w.linef(`m.%v = msg.Uint64(%d)`, name, tag)

	case compiler.KindFloat32:
		w.linef(`m.%v = msg.Float32(%d)`, name, tag)
	case compiler.KindFloat64:
		w.linef(`m.%v = msg.Float64(%d)`, name, tag)

	case compiler.KindBytes:
		w.linef(`m.%v = msg.Bytes(%d)`, name, tag)
	case compiler.KindString:
		w.linef(`m.%v = msg.String(%d)`, name, tag)

	// element-base

	case compiler.KindNullable:
		// val := w.readFieldTag(typ.Element, field.Tag)
		// w.linef(`m.%v = %v`, name, val)

	case compiler.KindList:
		// elem := typ.Element
		// elemName := goTypeName(elem)

		// // begin
		// w.line()
		// w.line(`{`)
		// w.linef(`list := msg.List(%d)`, tag)
		// w.linef(`ln := list.Len()`)
		// w.linef(`m.%v = make([]%v, 0, ln)`, name, elemName)
		// w.line()

		// // elements
		// w.linef(`for i := 0; i < ln; i++ {`)
		// val := w.readElement(elem)
		// w.linef(`elem := %v`, val)
		// w.linef(`m.%v = append(m.%v, elem)`, name, name)
		// w.line(`}`)

		// // end
		// w.line(`}`)
		// w.line()

		// references

	case compiler.KindEnum:
		typeName := goTypeName(typ)
		w.linef(`m.%v = %v(msg.Int32(%d))`, name, typeName, tag)

	case compiler.KindMessage:
		readFunc := ""
		if typ.ImportName == "" {
			readFunc = fmt.Sprintf("Read%vValue", typ.Name)
		} else {
			readFunc = fmt.Sprintf("%v.Read%vValue", typ.ImportName, typ.Name)
		}

		w.linef(`m.%v, _ = %v(msg.Field(%d))`, name, readFunc, tag)

	case compiler.KindStruct:
		typeName := goTypeName(typ)
		w.linef(`m.%v = Read%v(msg.Field(%d))`, name, typeName, tag)
	}
	return nil
}

// read struct

func (w *goWriter) readStruct(def *compiler.Definition) error {
	return nil
}

// read element

func (w *goWriter) readElement(typ *compiler.Type) string {
	kind := typ.Kind

	switch kind {
	case compiler.KindBool:
		return fmt.Sprintf(`list.Bool(i)`)

	case compiler.KindInt8:
		return fmt.Sprintf(`list.Int8(i)`)
	case compiler.KindInt16:
		return fmt.Sprintf(`list.Int16(i)`)
	case compiler.KindInt32:
		return fmt.Sprintf(`list.Int32(i)`)
	case compiler.KindInt64:
		return fmt.Sprintf(`list.Int64(i)`)

	case compiler.KindUint8:
		return fmt.Sprintf(`list.Uint8(i)`)
	case compiler.KindUint16:
		return fmt.Sprintf(`list.Uint16(i)`)
	case compiler.KindUint32:
		return fmt.Sprintf(`list.Uint32(i)`)
	case compiler.KindUint64:
		return fmt.Sprintf(`list.Uint64(i)`)

	case compiler.KindFloat32:
		return fmt.Sprintf(`list.Float32(i)`)
	case compiler.KindFloat64:
		return fmt.Sprintf(`list.Float64(i)`)

	case compiler.KindBytes:
		return fmt.Sprintf(`list.Bytes(i)`)
	case compiler.KindString:
		return fmt.Sprintf(`list.String(i)`)

	// element-based

	case compiler.KindList:
		panic("cannot read list as list element")
	case compiler.KindNullable:
		panic("cannot read nullable as list element")

	// references

	case compiler.KindEnum:
		typeName := goTypeName(typ)
		return fmt.Sprintf(`%v(list.Int32(i))`, typeName)

	case compiler.KindMessage:
		typeName := goTypeName(typ)
		w.linef(`data := list.Element(i)`)
		w.linef(`tmp, err := Read%v(data)
		if err != nil {
			return err
		}`, typeName)
		return `tmp`

	case compiler.KindStruct:
		typeName := goTypeName(typ)
		w.linef(`data := list.Element(i)`)
		w.linef(`tmp, err := Read%v(data)
		if err != nil {
			return err
		}`, typeName)
		return `tmp`
	}

	panic(fmt.Sprintf("unsupported type kind %v", kind))
}

// writes

func (w *goWriter) writes(file *compiler.File) error {
	w.line("// Write")
	w.line()

	for _, def := range file.Definitions {
		switch def.Type {
		case compiler.DefinitionEnum:
			// enum has no write method

		case compiler.DefinitionMessage:
			if err := w.writeMessage(def); err != nil {
				return err
			}
		case compiler.DefinitionStruct:
			if err := w.writeStruct(def); err != nil {
				return err
			}
		}
	}
	return nil
}

// write message

func (w *goWriter) writeMessage(def *compiler.Definition) error {
	w.linef(`func (m *%v) Write(w spec.Writer) error {`, def.Name)
	w.line(`if err := w.BeginMessage(); err != nil {
		return err
	}`)
	w.line()

	for _, field := range def.Message.Fields {
		if err := w.writeMessageField(field); err != nil {
			return err
		}
	}

	w.line(`return w.EndMessage()`)
	w.line("}")
	return nil
}

func (w *goWriter) writeMessageField(field *compiler.MessageField) error {
	name := goMessageFieldName(field)
	kind := field.Type.Kind

	typ := field.Type
	val := fmt.Sprintf("m.%v", name)

	switch kind {
	default:
		panic(fmt.Sprintf("unsupported type kind %v", kind))

	case compiler.KindBool:
		w.writeValue(typ, val)
		w.linef(`w.Field(%d)`, field.Tag)

	case compiler.KindInt8,
		compiler.KindInt16,
		compiler.KindInt32,
		compiler.KindInt64,

		compiler.KindUint8,
		compiler.KindUint16,
		compiler.KindUint32,
		compiler.KindUint64,

		compiler.KindFloat32,
		compiler.KindFloat64:
		w.linef(`if %v != 0 {`, val)
		w.writeValue(typ, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case compiler.KindBytes:
		w.linef(`if len(%v) > 0 {`, val)
		w.writeValue(typ, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case compiler.KindString:
		w.linef(`if len(%v) > 0 {`, val)
		w.writeValue(typ, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	// element-base

	case compiler.KindNullable:
		elem := typ.Element
		elemVal := "(*" + val + ")"

		w.linef(`if %v != nil {`, val)
		w.writeValue(elem, elemVal)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case compiler.KindList:
		elem := typ.Element

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

	// references

	case compiler.KindEnum:
		w.linef(`if %v != 0 {`, val)
		w.writeValue(typ, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case compiler.KindMessage:
		w.writeValue(typ, val)
		w.linef(`w.Field(%d)`, field.Tag)

	case compiler.KindStruct:
		w.writeValue(typ, val)
		w.linef(`w.Field(%d)`, field.Tag)

	}

	w.line()
	return nil
}

// write struct

func (w *goWriter) writeStruct(def *compiler.Definition) error {
	return nil
}

// write value

func (w *goWriter) writeValue(t *compiler.Type, val string) error {
	switch t.Kind {
	default:
		panic(fmt.Sprintf("unsupported type %v", t.Kind))

	// builtin

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

	// element-based

	case compiler.KindList:
		panic("cannot write list as value, write elements instead")
	case compiler.KindNullable:
		panic("cannot write nullable as value, write element instead")

	// references

	case compiler.KindEnum:
		w.linef(`w.Int32(int32(%v))`, val)
	case compiler.KindMessage:
		w.linef(`if err := %v.Write(w); err != nil { return err }`, val)
	case compiler.KindStruct:
		w.linef(`if err := %v.Write(w); err != nil { return err }`, val)
	}
	return nil
}

// struct

func (w *goWriter) struct_(def *compiler.Definition) error {
	return nil
}

// utils

func goPackageName(pkg *compiler.Package) string {
	return pkg.Name
}

func goImportName(imp *compiler.Import) string {
	id := imp.ID
	return id
}

func goEnumValueName(val *compiler.EnumValue) string {
	name := toUpperCamelCase(val.Name)
	return val.Enum.Def.Name + name
}

func goMessageFieldName(field *compiler.MessageField) string {
	return toUpperCamelCase(field.Name)
}

func goStructFieldName(field *compiler.StructField) string {
	return toUpperCamelCase(field.Name)
}

func goTypeName(t *compiler.Type) string {
	switch t.Kind {
	case compiler.KindBool:
		return "bool"

	case compiler.KindInt8:
		return "int8"
	case compiler.KindInt16:
		return "int16"
	case compiler.KindInt32:
		return "int32"
	case compiler.KindInt64:
		return "int64"

	case compiler.KindUint8:
		return "uint8"
	case compiler.KindUint16:
		return "uint16"
	case compiler.KindUint32:
		return "uint32"
	case compiler.KindUint64:
		return "uint64"

	case compiler.KindFloat32:
		return "float32"
	case compiler.KindFloat64:
		return "float64"

	case compiler.KindBytes:
		return "[]byte"
	case compiler.KindString:
		return "string"

	// element-based

	case compiler.KindList:
		elem := goTypeName(t.Element)
		return "[]" + elem
	case compiler.KindNullable:
		elem := goTypeName(t.Element)
		return "*" + elem

	// references

	case compiler.KindEnum,
		compiler.KindMessage,
		compiler.KindStruct:
		if t.Import != nil {
			return t.ImportName + "." + t.Name
		}
		return t.Name
	}

	return ""
}
