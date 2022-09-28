package golang

import (
	"fmt"

	"github.com/complex1tech/spec/lang/compiler"
)

func (w *writer) message(def *compiler.Definition) error {
	if err := w.messageDef(def); err != nil {
		return err
	}
	if err := w.newMessage(def); err != nil {
		return err
	}
	if err := w.decodeMessage(def); err != nil {
		return err
	}
	if err := w.messageFields(def); err != nil {
		return err
	}
	if err := w.messageHasFields(def); err != nil {
		return err
	}
	if err := w.messageMethods(def); err != nil {
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

func (w *writer) newMessage(def *compiler.Definition) error {
	w.linef(`func New%v(b []byte) %v {`, def.Name, def.Name)
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
		w.writef(`func (m %v) %v() %v {`, def.Name, fieldName, typeName)

		switch kind {
		case compiler.KindBool:
			w.writef(`return m.msg.GetBool(%d)`, tag)
		case compiler.KindByte:
			w.writef(`return m.msg.GetByte(%d)`, tag)

		case compiler.KindInt32:
			w.writef(`return m.msg.GetInt32(%d)`, tag)
		case compiler.KindInt64:
			w.writef(`return m.msg.GetInt64(%d)`, tag)
		case compiler.KindUint32:
			w.writef(`return m.msg.GetUint32(%d)`, tag)
		case compiler.KindUint64:
			w.writef(`return m.msg.GetUint64(%d)`, tag)

		case compiler.KindBin128:
			w.writef(`return m.msg.GetBin128(%d)`, tag)
		case compiler.KindBin256:
			w.writef(`return m.msg.GetBin256(%d)`, tag)

		case compiler.KindFloat32:
			w.writef(`return m.msg.GetFloat32(%d)`, tag)
		case compiler.KindFloat64:
			w.writef(`return m.msg.GetFloat64(%d)`, tag)

		case compiler.KindBytes:
			w.writef(`return m.msg.GetBytes(%d)`, tag)
		case compiler.KindString:
			w.writef(`return m.msg.GetString(%d)`, tag)
		}

		w.writef(`}`)
		w.line()

	case compiler.KindList:
		decodeFunc := typeDecodeFunc(field.Type.Element)

		w.writef(`func (m %v) %v() %v {`, def.Name, fieldName, typeName)
		w.writef(`return spec.NewList(m.msg.Field(%d), %v)`, tag, decodeFunc)
		w.writef(`}`)
		w.line()

	case compiler.KindEnum,
		compiler.KindMessage,
		compiler.KindStruct:
		newFunc := typeNewFunc(field.Type)

		w.writef(`func (m %v) %v() %v {`, def.Name, fieldName, typeName)
		w.writef(`return %v(m.msg.Field(%d))`, newFunc, tag)
		w.writef(`}`)
		w.line()
	}
	return nil
}

func (w *writer) messageHasFields(def *compiler.Definition) error {
	for _, field := range def.Message.Fields {
		if err := w.messageHasField(def, field); err != nil {
			return err
		}
	}

	w.line()
	return nil
}

func (w *writer) messageHasField(def *compiler.Definition, field *compiler.MessageField) error {
	fieldName := messageFieldName(field)
	tag := field.Tag

	w.writef(`func (m %v) Has%v() bool {`, def.Name, fieldName)
	w.writef(`return m.msg.HasField(%d)`, tag)
	w.writef(`}`)
	w.line()
	return nil
}

func (w *writer) messageMethods(def *compiler.Definition) error {
	w.writef(`func (m %v) IsEmpty() bool {`, def.Name)
	w.writef(`return m.msg.Empty()`)
	w.writef(`}`)
	w.line()

	w.writef(`func (m %v) Clone() %v {`, def.Name, def.Name)
	w.writef(`return %v{m.msg.Clone()}`, def.Name)
	w.writef(`}`)
	w.line()

	w.writef(`func (m %v) Unwrap() spec.Message {`, def.Name)
	w.writef(`return m.msg`)
	w.writef(`}`)
	w.line()
	return nil
}

// builder

func (w *writer) messageBuilder(def *compiler.Definition) error {
	if err := w.messageBuilderDef(def); err != nil {
		return err
	}
	if err := w.newMessageBuilder(def); err != nil {
		return err
	}
	if err := w.messageBuilderFields(def); err != nil {
		return err
	}
	if err := w.messageBuilderBuild(def); err != nil {
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

func (w *writer) newMessageBuilder(def *compiler.Definition) error {
	w.linef(`func New%vBuilder() %vBuilder {`, def.Name, def.Name)
	w.linef(`e := spec.NewEncoder()`)
	w.linef(`e.BeginMessage()`)
	w.linef(`return %vBuilder{e}`, def.Name)
	w.linef(`}`)
	w.line()

	w.linef(`func New%vBuilderBuffer(b buffer.Buffer) %vBuilder {`, def.Name, def.Name)
	w.linef(`e := spec.NewEncoderBuffer(b)`)
	w.linef(`e.BeginMessage()`)
	w.linef(`return %vBuilder{e}`, def.Name)
	w.linef(`}`)
	w.line()

	w.linef(`func New%vBuilderEncoder(e *spec.Encoder) %vBuilder {`, def.Name, def.Name)
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
	w.linef(`return New%v(bytes), nil`, def.Name)
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
	bname := fmt.Sprintf("%vBuilder", def.Name)

	tag := field.Tag
	kind := field.Type.Kind

	switch kind {
	default:
		w.linef(`func (b %vBuilder) %v(v %v) %v {`, def.Name, fname, tname, bname)

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

		w.linef(`b.e.Field(%d)`, tag)
		w.linef(`return b`)
		w.linef(`}`)
		w.line()

	case compiler.KindEnum:
		encodeFunc := typeEncodeFunc(field.Type)

		w.linef(`func (b %v) %v(v %v) %v {`, bname, fname, tname, bname)
		w.linef(`spec.EncodeValue(b.e, v, %v)`, encodeFunc)
		w.linef(`b.e.Field(%d)`, tag)
		w.linef(`return b`)
		w.linef(`}`)
		w.line()

	case compiler.KindStruct:
		encodeFunc := typeEncodeFunc(field.Type)

		w.linef(`func (b %v) %v(v %v) %v {`, bname, fname, tname, bname)
		w.linef(`spec.EncodeValue(b.e, v, %v)`, encodeFunc)
		w.linef(`b.e.Field(%d)`, tag)
		w.linef(`return b`)
		w.linef(`}`)
		w.line()

	case compiler.KindList:
		builder := typeBuilder(field.Type)
		buildList := typeEncodeFunc(field.Type)
		encodeElement := typeEncodeFunc(field.Type.Element)

		w.linef(`func (b %v) %v() %v {`, bname, fname, builder)
		w.linef(`b.e.BeginField(%d)`, tag)
		w.linef(`return %v(b.e, %v)`, buildList, encodeElement)
		w.linef(`}`)
		w.line()

	case compiler.KindMessage:
		builder := typeBuilder(field.Type)
		newMessageBuilder := typeEncodeFunc(field.Type)
		w.linef(`func (b %v) %v() %v {`, bname, fname, builder)
		w.linef(`b.e.BeginField(%d)`, tag)
		w.linef(`return %v(b.e)`, newMessageBuilder)
		w.linef(`}`)
		w.line()

		tname := typeName(field.Type)
		w.linef(`func (b %v) Copy%v(v %v) error {`, bname, fname, tname)
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
