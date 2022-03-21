package golang

import (
	"github.com/complexl/spec/lang/compiler"
)

func (w *writer) message(def *compiler.Definition) error {
	if err := w.messageDef(def); err != nil {
		return err
	}
	if err := w.getMessage(def); err != nil {
		return err
	}
	if err := w.decodeMessage(def); err != nil {
		return err
	}
	if err := w.messageFields(def); err != nil {
		return err
	}
	if err := w.messageRawBytes(def); err != nil {
		return err
	}
	return nil
}

func (w *writer) messageDef(def *compiler.Definition) error {
	w.linef(`// %v`, def.Name)
	w.line()
	w.linef(`type %v struct {`, def.Name)
	w.line(`msg spec.Message`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) getMessage(def *compiler.Definition) error {
	w.linef(`func Get%v(b []byte) %v {`, def.Name, def.Name)
	w.linef(`msg := spec.GetMessage(b)`)
	w.linef(`return %v{msg}`, def.Name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) decodeMessage(def *compiler.Definition) error {
	w.linef(`func Decode%v(b []byte) (result %v, size int, err error) {`, def.Name, def.Name)
	w.linef(`msg, size, err := spec.DecodeMessage(b)`)
	w.line(`if err != nil {
		return
	}`)
	w.linef(`result = %v{msg}`, def.Name)
	w.linef(`return`)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) messageRawBytes(def *compiler.Definition) error {
	w.linef(`func (m %v) RawBytes() []byte {`, def.Name)
	w.linef(`return m.msg.Raw()`)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) messageFields(def *compiler.Definition) error {
	for _, field := range def.Message.Fields {
		if err := w.messageField(def, field); err != nil {
			return err
		}
	}

	w.line()
	return nil
}

func (w *writer) messageField(def *compiler.Definition, field *compiler.MessageField) error {
	fieldName := messageFieldName(field)
	typeName := typeName(field.Type)

	tag := field.Tag
	kind := field.Type.Kind

	switch kind {
	default:
		w.linef(`func (m %v) %v() %v {`, def.Name, fieldName, typeName)

		switch kind {
		case compiler.KindBool:
			w.linef(`return m.msg.Bool(%d)`, tag)
		case compiler.KindByte:
			w.linef(`return m.msg.Byte(%d)`, tag)

		case compiler.KindInt32:
			w.linef(`return m.msg.Int32(%d)`, tag)
		case compiler.KindInt64:
			w.linef(`return m.msg.Int64(%d)`, tag)
		case compiler.KindUint32:
			w.linef(`return m.msg.Uint32(%d)`, tag)
		case compiler.KindUint64:
			w.linef(`return m.msg.Uint64(%d)`, tag)

		case compiler.KindU128:
			w.linef(`return m.msg.U128(%d)`, tag)
		case compiler.KindU256:
			w.linef(`return m.msg.U256(%d)`, tag)

		case compiler.KindFloat32:
			w.linef(`return m.msg.Float32(%d)`, tag)
		case compiler.KindFloat64:
			w.linef(`return m.msg.Float64(%d)`, tag)

		case compiler.KindBytes:
			w.linef(`return m.msg.Bytes(%d)`, tag)
		case compiler.KindString:
			w.linef(`return m.msg.String(%d)`, tag)
		}

		w.linef(`}`)
		w.line()

	case compiler.KindList:
		decodeFunc := typeDecodeFunc(field.Type.Element)

		w.linef(`func (m %v) %v() %v {`, def.Name, fieldName, typeName)
		w.linef(`b := m.msg.Field(%d)`, tag)
		w.linef(`return spec.GetList(b, %v)`, decodeFunc)
		w.linef(`}`)
		w.line()

	case compiler.KindEnum,
		compiler.KindMessage,
		compiler.KindStruct:
		getFunc := typeGetFunc(field.Type)

		w.linef(`func (m %v) %v() %v {`, def.Name, fieldName, typeName)
		w.linef(`b := m.msg.Field(%d)`, tag)
		w.linef(`return %v(b)`, getFunc)
		w.linef(`}`)
		w.line()
	}
	return nil
}

// util

func messageFieldName(field *compiler.MessageField) string {
	return toUpperCamelCase(field.Name)
}
