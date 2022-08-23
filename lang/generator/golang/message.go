package golang

import (
	"github.com/epochtimeout/spec/lang/compiler"
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
	if err := w.messageMethods(def); err != nil {
		return err
	}
	if err := w.messageFields(def); err != nil {
		return err
	}

	if err := w.messageBuilder(def); err != nil {
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
	w.linef(`func Decode%v(b []byte) (_ %v, size int, err error) {`, def.Name, def.Name)
	w.linef(`msg, size, err := spec.DecodeMessage(b)`)
	w.line(`if err != nil || size == 0 {
		return
	}`)
	w.linef(`return %v{msg}, size, nil`, def.Name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) messageMethods(def *compiler.Definition) error {
	w.linef(`func (m %v) Clone() %v {`, def.Name, def.Name)
	w.linef(`msg1 := m.msg.Clone()`)
	w.linef(`return %v{msg1}`, def.Name)
	w.linef(`}`)
	w.line()

	w.linef(`func (m %v) IsEmpty() bool {`, def.Name)
	w.linef(`return m.msg.IsEmpty()`)
	w.linef(`}`)
	w.line()

	w.linef(`func (m %v) Unwrap() spec.Message {`, def.Name)
	w.linef(`return m.msg`)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) messageFields(def *compiler.Definition) error {
	for _, field := range def.Message.Fields {
		if err := w.messageField(def, field); err != nil {
			return err
		}
		if err := w.messageHasField(def, field); err != nil {
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
			w.linef(`return m.msg.GetBool(%d)`, tag)
		case compiler.KindByte:
			w.linef(`return m.msg.GetByte(%d)`, tag)

		case compiler.KindInt32:
			w.linef(`return m.msg.GetInt32(%d)`, tag)
		case compiler.KindInt64:
			w.linef(`return m.msg.GetInt64(%d)`, tag)
		case compiler.KindUint32:
			w.linef(`return m.msg.GetUint32(%d)`, tag)
		case compiler.KindUint64:
			w.linef(`return m.msg.GetUint64(%d)`, tag)

		case compiler.KindBin128:
			w.linef(`return m.msg.GetBin128(%d)`, tag)
		case compiler.KindBin256:
			w.linef(`return m.msg.GetBin256(%d)`, tag)

		case compiler.KindFloat32:
			w.linef(`return m.msg.GetFloat32(%d)`, tag)
		case compiler.KindFloat64:
			w.linef(`return m.msg.GetFloat64(%d)`, tag)

		case compiler.KindBytes:
			w.linef(`return m.msg.GetBytes(%d)`, tag)
		case compiler.KindString:
			w.linef(`return m.msg.GetString(%d)`, tag)
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

func (w *writer) messageHasField(def *compiler.Definition, field *compiler.MessageField) error {
	fieldName := messageFieldName(field)
	tag := field.Tag

	w.linef(`func (m %v) Has%v() bool {`, def.Name, fieldName)
	w.linef(`return m.msg.HasField(%d)`, tag)
	w.linef(`}`)
	w.line()
	return nil
}

// encoder

func (w *writer) messageBuilder(def *compiler.Definition) error {
	if err := w.messageBuilderDef(def); err != nil {
		return err
	}
	if err := w.buildMessage(def); err != nil {
		return err
	}
	if err := w.messageBuilderBuild(def); err != nil {
		return err
	}
	if err := w.messageBuilderFields(def); err != nil {
		return err
	}
	return nil
}

func (w *writer) messageBuilderDef(def *compiler.Definition) error {
	w.linef(`// %vBuilder`, def.Name)
	w.line()
	w.linef(`type %vBuilder struct {`, def.Name)
	w.line(`e *spec.Encoder`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) buildMessage(def *compiler.Definition) error {
	w.linef(`func Build%v() %vBuilder {`, def.Name, def.Name)
	w.linef(`e := spec.NewEncoder()`)
	w.linef(`e.BeginMessage()`)
	w.linef(`return %vBuilder{e}`, def.Name)
	w.linef(`}`)
	w.line()

	w.linef(`func Build%vBuffer(b buffer.Buffer) %vBuilder {`, def.Name, def.Name)
	w.linef(`e := spec.NewEncoderBuffer(b)`)
	w.linef(`e.BeginMessage()`)
	w.linef(`return %vBuilder{e}`, def.Name)
	w.linef(`}`)
	w.line()

	w.linef(`func Build%vEncoder(e *spec.Encoder) %vBuilder {`, def.Name, def.Name)
	w.linef(`e.BeginMessage()`)
	w.linef(`return %vBuilder{e}`, def.Name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) messageBuilderBuild(def *compiler.Definition) error {
	w.linef(`func (b %vBuilder) End() error {`, def.Name)
	w.linef(`_, err := b.e.End()`)
	w.linef(`return err`)
	w.linef(`}`)
	w.line()

	w.linef(`func (b %vBuilder) Build() (_ %v, err error) {`, def.Name, def.Name)
	w.linef(`bytes, err := b.e.End()`)
	w.linef(`if err != nil {
		return
	}`)
	w.linef(`return Get%v(bytes), nil`, def.Name)
	w.linef(`}`)
	w.line()

	w.linef(`func (b %vBuilder) Unwrap() *spec.Encoder {`, def.Name)
	w.linef(`return b.e`)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) messageBuilderFields(def *compiler.Definition) error {
	for _, field := range def.Message.Fields {
		if err := w.messageBuilderField(def, field); err != nil {
			return err
		}
	}

	w.line()
	return nil
}

func (w *writer) messageBuilderField(def *compiler.Definition, field *compiler.MessageField) error {
	fname := messageFieldName(field)
	tname := typeName(field.Type)

	tag := field.Tag
	kind := field.Type.Kind

	switch kind {
	default:
		w.linef(`func (b %vBuilder) %v(v %v) error {`, def.Name, fname, tname)

		switch kind {
		case compiler.KindBool:
			w.line(`b.e.Bool(v)`)
		case compiler.KindByte:
			w.line(`b.e.Byte(v)`)

		case compiler.KindInt32:
			w.line(`b.e.Int32(v)`)
		case compiler.KindInt64:
			w.line(`b.e.Int64(v)`)
		case compiler.KindUint32:
			w.line(`b.e.Uint32(v)`)
		case compiler.KindUint64:
			w.line(`b.e.Uint64(v)`)

		case compiler.KindBin128:
			w.line(`b.e.Bin128(v)`)
		case compiler.KindBin256:
			w.line(`b.e.Bin256(v)`)

		case compiler.KindFloat32:
			w.line(`b.e.Float32(v)`)
		case compiler.KindFloat64:
			w.line(`b.e.Float64(v)`)

		case compiler.KindBytes:
			w.line(`b.e.Bytes(v)`)
		case compiler.KindString:
			w.line(`b.e.String(v)`)
		}

		w.linef(`return b.e.Field(%d)`, tag)
		w.linef(`}`)
		w.line()

	case compiler.KindEnum:
		encodeFunc := typeEncodeFunc(field.Type)

		w.linef(`func (b %vBuilder) %v(v %v) error {`, def.Name, fname, tname)
		w.linef(`spec.EncodeValue(b.e, v, %v)`, encodeFunc)
		w.linef(`return b.e.Field(%d)`, tag)
		w.linef(`}`)
		w.line()

	case compiler.KindStruct:
		encodeFunc := typeEncodeFunc(field.Type)

		w.linef(`func (b %vBuilder) %v(v %v) error {`, def.Name, fname, tname)
		w.linef(`spec.EncodeValue(b.e, v, %v)`, encodeFunc)
		w.linef(`return b.e.Field(%d)`, tag)
		w.linef(`}`)
		w.line()

	case compiler.KindList:
		builder := typeBuilder(field.Type)
		buildList := typeEncodeFunc(field.Type)
		encodeElement := typeEncodeFunc(field.Type.Element)

		w.linef(`func (b %vBuilder) %v() %v {`, def.Name, fname, builder)
		w.linef(`b.e.BeginField(%d)`, tag)
		w.linef(`return %v(b.e, %v)`, buildList, encodeElement)
		w.linef(`}`)
		w.line()

	case compiler.KindMessage:
		builder := typeBuilder(field.Type)
		buildMessage := typeEncodeFunc(field.Type)
		w.linef(`func (b %vBuilder) %v() %v {`, def.Name, fname, builder)
		w.linef(`b.e.BeginField(%d)`, tag)
		w.linef(`return %v(b.e)`, buildMessage)
		w.linef(`}`)
		w.line()

		tname := typeName(field.Type)
		w.linef(`func (b %vBuilder) Copy%v(v %v) error {`, def.Name, fname, tname)
		w.linef(`return b.e.FieldBytes(%d, v.Unwrap().Bytes())`, tag)
		w.linef(`}`)
		w.line()
	}
	return nil
}

// util

func messageFieldName(field *compiler.MessageField) string {
	return toUpperCamelCase(field.Name)
}
