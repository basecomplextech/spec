package golang

import (
	"fmt"

	"github.com/baseone-run/spec/compiler"
)

func (w *writer) message(def *compiler.Definition) error {
	// message
	if err := w.messageDef(def); err != nil {
		return err
	}
	if err := w.readMessage(def); err != nil {
		return err
	}
	if err := w.messageRead(def); err != nil {
		return err
	}
	if err := w.messageWrite(def); err != nil {
		return err
	}

	// data
	if err := w.messageData(def); err != nil {
		return err
	}
	return nil
}

func (w *writer) messageDef(def *compiler.Definition) error {
	w.linef(`// %v`, def.Name)
	w.line()
	w.linef("type %v struct {", def.Name)

	for _, field := range def.Message.Fields {
		name := messageFieldName(field)
		typ := objectType(field.Type)
		tag := fmt.Sprintf("`tag:\"%d\" json:\"%v\"`", field.Tag, field.Name)
		w.linef("%v %v %v", name, typ, tag)
	}

	w.line("}")
	w.line()
	return nil
}

func (w *writer) readMessage(def *compiler.Definition) error {
	w.linef(`func Read%v(b []byte) (*%v, error) {`, def.Name, def.Name)
	w.linef(`if len(b) == 0 {
		return nil, nil
	}`)
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
	w.line(`r, err := spec.NewMessageReader(b)
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
	tag := fmt.Sprintf("%d", field.Tag)

	switch kind {
	default:
		stmt := w.readerRead(typ, "r", tag)
		w.linef(`m.%v, err = %v`, name, stmt)
		w.linef(`if err != nil {
			return err
		}`)

	case compiler.KindList,
		compiler.KindEnum,
		compiler.KindMessage,
		compiler.KindStruct:
		// wrap in {}

		w.linef(`{`)
		stmt := w.readerRead(typ, "r", tag)
		w.linef(`m.%v, err = %v`, name, stmt)
		w.linef(`if err != nil {
			return err
		}`)

		w.linef(`}`)
	}
	return nil
}

func (w *writer) messageWrite(def *compiler.Definition) error {
	// begin
	w.linef(`func (m *%v) Write(w *spec.Writer) error {`, def.Name)
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
		w.writerWrite(typ, val)
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
		// numbers

		w.linef(`if %v != 0 {`, val)
		w.writerWrite(typ, val)
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
		if err := w.writerWrite(elem, "elem"); err != nil {
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

	// resolvable

	case compiler.KindEnum:
		w.linef(`if %v != 0 {`, val)
		w.writerWrite(typ, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case compiler.KindMessage:
		w.linef(`if %v != nil {`, val)
		w.writerWrite(typ, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)

	case compiler.KindStruct:
		w.linef(`{`)
		w.writerWrite(typ, val)
		w.linef(`w.Field(%d)`, field.Tag)
		w.line(`}`)
	}
	return nil
}

// data

func (w *writer) messageData(def *compiler.Definition) error {
	if err := w.messageDataDef(def); err != nil {
		return err
	}
	if err := w.newMessageData(def); err != nil {
		return err
	}
	if err := w.readMessageData(def); err != nil {
		return err
	}
	if err := w.messageDataMethods(def); err != nil {
		return err
	}
	return nil
}

func (w *writer) messageDataDef(def *compiler.Definition) error {
	name := fmt.Sprintf("%vData", def.Name)

	w.linef(`// %v`, name)
	w.line()
	w.linef(`type %v struct {`, name)
	w.line(`d spec.MessageData`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) newMessageData(def *compiler.Definition) error {
	name := fmt.Sprintf("%vData", def.Name)

	w.linef(`func New%v(b []byte) (%v, error) {`, name, name)
	w.linef(`d, err := spec.NewMessageData(b)`)
	w.linef(`if err != nil {
		return %v{}, err
	}`, name)
	w.linef(`return %v{d}, nil`, name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) readMessageData(def *compiler.Definition) error {
	name := fmt.Sprintf("%vData", def.Name)

	w.linef(`func Read%v(b []byte) (%v, error) {`, name, name)
	w.linef(`d, err := spec.ReadMessageData(b)`)
	w.linef(`if err != nil {
		return %v{}, err
	}`, name)
	w.linef(`return %v{d}, nil`, name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) messageDataMethods(def *compiler.Definition) error {
	for _, field := range def.Message.Fields {
		if err := w.messageDataMethod(def, field); err != nil {
			return err
		}
	}

	w.line()
	return nil
}

func (w *writer) messageDataMethod(def *compiler.Definition, field *compiler.MessageField) error {
	name := messageFieldName(field)
	typ := field.Type
	tag := fmt.Sprintf("%d", field.Tag)
	kind := field.Type.Kind

	switch kind {
	default:
		dataType := dataType(field.Type)
		w.linef(`func (d %vData) %v() %v {`, def.Name, name, dataType)
		dataGet := w.dataGet(field.Type, "d.d", tag)
		w.linef(`return %v`, dataGet)
		w.linef(`}`)
		w.line()

	case compiler.KindList:
		elem := typ.Element
		elemName := dataType(elem)

		// len
		w.linef(`func (d %vData) %vLen() int {`, def.Name, name)
		w.linef(`return d.d.List(%v).Len()`, tag)
		w.linef(`}`)
		w.line()

		// element
		w.linef(`func (d %vData) %vElement(i int) %v {`, def.Name, name, elemName)
		w.linef(`data := d.d.List(%v).Element(i)`, tag)
		dataGet := w.dataGet(elem, "data", "")
		w.linef(`return %v`, dataGet)
		w.linef(`}`)
		w.line()

	case compiler.KindEnum,
		compiler.KindMessage,
		compiler.KindStruct:

		dataType := dataType(field.Type)
		w.linef(`func (d %vData) %v() %v {`, def.Name, name, dataType)
		w.linef(`data := d.d.Element(%v)`, tag)
		dataGet := w.dataGet(field.Type, "data", "")
		w.linef(`return %v`, dataGet)
		w.linef(`}`)
		w.line()
	}
	return nil
}

// util

func messageFieldName(field *compiler.MessageField) string {
	return toUpperCamelCase(field.Name)
}

func messageReadFunc(typ *compiler.Type) string {
	if typ.Import == nil {
		return fmt.Sprintf("Read%v", typ.Name)
	}
	return fmt.Sprintf("%v.Read%v", typ.ImportName, typ.Name)
}

func messageNewDataFunc(typ *compiler.Type) string {
	if typ.Import == nil {
		return fmt.Sprintf("New%vData", typ.Name)
	}
	return fmt.Sprintf("%v.New%vData", typ.ImportName, typ.Name)
}

func messageReadDataFunc(typ *compiler.Type) string {
	if typ.Import == nil {
		return fmt.Sprintf("Read%vData", typ.Name)
	}
	return fmt.Sprintf("%v.Read%vData", typ.ImportName, typ.Name)
}
