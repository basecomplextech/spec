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
	if err := w.encodeMessage(def); err != nil {
		return err
	}
	if err := w.messageFields(def); err != nil {
		return err
	}
	if err := w.messageRawBytes(def); err != nil {
		return err
	}
	if err := w.messageEncoder(def); err != nil {
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

func (w *writer) encodeMessage(def *compiler.Definition) error {
	w.linef(`func Encode%v(e *spec.Encoder) (result %vEncoder, err error) {`, def.Name, def.Name)
	w.linef(`if err = e.BeginMessage(); err != nil {
		return
	}`)
	w.linef(`result = %vEncoder{e}`, def.Name)
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

// encoder

func (w *writer) messageEncoder(def *compiler.Definition) error {
	if err := w.messageEncoderDef(def); err != nil {
		return err
	}
	if err := w.messageEncoderEnd(def); err != nil {
		return err
	}
	if err := w.messageEncoderFields(def); err != nil {
		return err
	}
	return nil
}

func (w *writer) messageEncoderDef(def *compiler.Definition) error {
	w.linef(`// %vEncoder`, def.Name)
	w.line()
	w.linef(`type %vEncoder struct {`, def.Name)
	w.line(`e *spec.Encoder`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) messageEncoderEnd(def *compiler.Definition) error {
	w.linef(`func (e %vEncoder) End() ([]byte, error) {`, def.Name)
	w.linef(`return e.e.End()`)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) messageEncoderFields(def *compiler.Definition) error {
	for _, field := range def.Message.Fields {
		if err := w.messageEncoderField(def, field); err != nil {
			return err
		}
	}

	w.line()
	return nil
}

func (w *writer) messageEncoderField(def *compiler.Definition, field *compiler.MessageField) error {
	fieldName := messageFieldName(field)
	typeName := typeName(field.Type)

	tag := field.Tag
	kind := field.Type.Kind

	switch kind {
	default:
		w.linef(`func (e %vEncoder) %v(v %v) error {`, def.Name, fieldName, typeName)

		switch kind {
		case compiler.KindBool:
			w.line(`e.e.Bool(v)`)
		case compiler.KindByte:
			w.line(`e.e.Byte(v)`)

		case compiler.KindInt32:
			w.line(`e.e.Int32(v)`)
		case compiler.KindInt64:
			w.line(`e.e.Int64(v)`)
		case compiler.KindUint32:
			w.line(`e.e.Uint32(v)`)
		case compiler.KindUint64:
			w.line(`e.e.Uint64(v)`)

		case compiler.KindU128:
			w.line(`e.e.U128(v)`)
		case compiler.KindU256:
			w.line(`e.e.U256(v)`)

		case compiler.KindFloat32:
			w.line(`e.e.Float32(v)`)
		case compiler.KindFloat64:
			w.line(`e.e.Float64(v)`)

		case compiler.KindBytes:
			w.line(`e.e.Bytes(v)`)
		case compiler.KindString:
			w.line(`e.e.String(v)`)
		}

		w.linef(`return e.e.Field(%d)`, tag)
		w.linef(`}`)
		w.line()

	case compiler.KindEnum:
		encodeFunc := typeEncodeFunc(field.Type)

		w.linef(`func (e %vEncoder) %v(v %v) error {`, def.Name, fieldName, typeName)
		w.linef(`%v(e.e, v)`, encodeFunc)
		w.linef(`return e.e.Field(%d)`, tag)
		w.linef(`}`)
		w.line()

	case compiler.KindStruct:
		encodeFunc := typeEncodeFunc(field.Type)

		w.linef(`func (e %vEncoder) %v(v %v) error {`, def.Name, fieldName, typeName)
		w.linef(`spec.EncodeValue(e.e, v, %v)`, encodeFunc)
		w.linef(`return e.e.Field(%d)`, tag)
		w.linef(`}`)
		w.line()

	case compiler.KindList:
		encoder := typeEncoder(field.Type)
		encodeFunc := typeEncodeFunc(field.Type)
		encodeElemFunc := typeEncodeFunc(field.Type.Element)

		w.linef(`func (e %vEncoder) %v() (%v, error) {`, def.Name, fieldName, encoder)
		w.linef(`e.e.BeginField(%d)`, tag)
		w.linef(`return %v(e.e, %v)`, encodeFunc, encodeElemFunc)
		w.linef(`}`)
		w.line()

	case compiler.KindMessage:
		encoder := typeEncoder(field.Type)
		encodeFunc := typeEncodeFunc(field.Type)

		w.linef(`func (e %vEncoder) %v() (%v, error) {`, def.Name, fieldName, encoder)
		w.linef(`e.e.BeginField(%d)`, tag)
		w.linef(`return %v(e.e)`, encodeFunc)
		w.linef(`}`)
		w.line()
	}
	return nil
}

// util

func messageFieldName(field *compiler.MessageField) string {
	return toUpperCamelCase(field.Name)
}
