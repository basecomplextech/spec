package golang

import (
	"fmt"

	"github.com/baseone-run/spec/compiler"
)

func (w *writer) message(def *compiler.Definition) error {
	if err := w.messageDef(def); err != nil {
		return err
	}
	if err := w.messageReadFunc(def); err != nil {
		return err
	}
	if err := w.messageRead(def); err != nil {
		return err
	}
	if err := w.messageWrite(def); err != nil {
		return err
	}
	return nil
}

// def

func (w *writer) messageDef(def *compiler.Definition) error {
	w.linef("type %v struct {", def.Name)

	for _, field := range def.Message.Fields {
		name := messageFieldName(field)
		type_ := typeName(field.Type)
		tag := fmt.Sprintf("`tag:\"%d\" json:\"%v\"`", field.Tag, field.Name)
		w.linef("%v %v %v", name, type_, tag)
	}

	w.line("}")
	w.line()
	return nil
}

// read

func (w *writer) messageReadFunc(def *compiler.Definition) error {
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

func (w *writer) messageRead(def *compiler.Definition) error {
	w.linef(`func (m *%v) Read(b []byte) error {`, def.Name)
	w.line(`msg, err := spec.ReadMessage(b)
	if err != nil {
		return err
	}`)
	w.line()

	for _, field := range def.Message.Fields {
		if err := w.messageReadField(field); err != nil {
			return err
		}
	}

	w.line(`return nil`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) messageReadField(field *compiler.MessageField) error {
	name := messageFieldName(field)
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

	// list

	case compiler.KindList:
		elem := typ.Element

		// begin
		w.line(`{`)
		w.linef(`list := msg.List(%d)`, tag)
		w.linef(`ln := list.Len()`)
		w.linef(`m.%v = make([]%v, 0, ln)`, name, typeName(elem))
		w.line()

		// elements
		w.linef(`for i := 0; i < ln; i++ {`)
		w.listReadElement(elem)
		w.linef(`m.%v = append(m.%v, elem)`, name, name)
		w.line(`}`)

		// end
		w.line(`}`)
		w.line()

	// resolved

	case compiler.KindEnum:
		typeName := typeName(typ)
		w.linef(`m.%v = %v(msg.Int32(%d))`, name, typeName, tag)

	case compiler.KindMessage:
		readFunc := typeReadFunc(typ)
		w.linef(`m.%v, _ = %v(msg.Field(%d))`, name, readFunc, tag)

	case compiler.KindStruct:
		readFunc := typeReadFunc(typ)
		w.linef(`m.%v, _ = %v(msg.Field(%d))`, name, readFunc, tag)
	}
	return nil
}

// write

func (w *writer) messageWrite(def *compiler.Definition) error {
	// begin
	w.linef(`func (m *%v) Write(w spec.Writer) error {`, def.Name)
	w.linef(`if m == nil {
		return w.Nil()
	}`)
	w.line(`if err := w.BeginMessage(); err != nil {
		return err
	}`)
	w.line()

	// fields
	for _, field := range def.Message.Fields {
		if err := w.messageWriteField(field); err != nil {
			return err
		}
	}

	// end
	w.line(`return w.EndMessage()`)
	w.line("}")
	return nil
}

func (w *writer) messageWriteField(field *compiler.MessageField) error {
	name := messageFieldName(field)
	kind := field.Type.Kind

	typ := field.Type
	val := fmt.Sprintf("m.%v", name)

	switch kind {
	default:
		panic(fmt.Sprintf("unsupported type kind %v", kind))

	case compiler.KindBool:
		w.line(`{`)
		w.writeValue(typ, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

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
		w.linef(`w.Bytes(%v)`, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case compiler.KindString:
		w.linef(`if len(%v) > 0 {`, val)
		w.linef(`w.String(%v)`, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	// list

	case compiler.KindList:
		elem := typ.Element

		// begin
		w.linef(`if len(%v) > 0 {`, val)
		w.line(`if err := w.BeginList(); err != nil {
			return err 
		}`)

		// elements
		w.linef(`for _, elem := range %v {`, val)
		if err := w.writeValue(elem, "elem"); err != nil {
			return err
		}
		w.line(`w.Element()`)
		w.line(`}`)

		// end
		w.line(`if err := w.EndList(); err != nil {
			return err
		}`)

		// field
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	// resolved

	case compiler.KindEnum:
		w.linef(`if %v != 0 {`, val)
		w.writeValue(typ, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case compiler.KindMessage:
		w.linef(`if %v != nil {`, val)
		w.writeValue(typ, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case compiler.KindStruct:
		w.line(`{`)
		w.linef(`if err := %v.Write(w); err != nil { return err }`, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)
	}
	return nil
}

func messageFieldName(field *compiler.MessageField) string {
	return toUpperCamelCase(field.Name)
}
