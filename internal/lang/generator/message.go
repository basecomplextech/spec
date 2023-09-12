package generator

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/model"
)

func (w *writer) message(def *model.Definition) error {
	if err := w.messageDef(def); err != nil {
		return err
	}
	if err := w.newMessage(def); err != nil {
		return err
	}
	if err := w.parseMessage(def); err != nil {
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

func (w *writer) messageDef(def *model.Definition) error {
	w.linef(`// %v`, def.Name)
	w.line()
	w.linef(`type %v struct {`, def.Name)
	w.line(`msg spec.Message`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) newMessage(def *model.Definition) error {
	w.linef(`func New%v(b []byte) %v {`, def.Name, def.Name)
	w.linef(`msg := spec.NewMessage(b)`)
	w.linef(`return %v{msg}`, def.Name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) parseMessage(def *model.Definition) error {
	w.linef(`func Parse%v(b []byte) (_ %v, size int, err error) {`, def.Name, def.Name)
	w.linef(`msg, size, err := spec.ParseMessage(b)`)
	w.line(`if err != nil || size == 0 {
		return
	}`)
	w.linef(`return %v{msg}, size, nil`, def.Name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) messageFields(def *model.Definition) error {
	fields := def.Message.Fields.List

	for _, field := range fields {
		if err := w.messageField(def, field); err != nil {
			return err
		}
	}

	w.line()
	return nil
}

func (w *writer) messageField(def *model.Definition, field *model.Field) error {
	fieldName := messageFieldName(field)
	typeName := typeRefName(field.Type)

	tag := field.Tag
	kind := field.Type.Kind

	switch kind {
	default:
		w.writef(`func (m %v) %v() %v {`, def.Name, fieldName, typeName)

		switch kind {
		case model.KindBool:
			w.writef(`return m.msg.Field(%d).Bool()`, tag)
		case model.KindByte:
			w.writef(`return m.msg.Field(%d).Byte()`, tag)

		case model.KindInt16:
			w.writef(`return m.msg.Field(%d).Int16()`, tag)
		case model.KindInt32:
			w.writef(`return m.msg.Field(%d).Int32()`, tag)
		case model.KindInt64:
			w.writef(`return m.msg.Field(%d).Int64()`, tag)

		case model.KindUint16:
			w.writef(`return m.msg.Field(%d).Uint16()`, tag)
		case model.KindUint32:
			w.writef(`return m.msg.Field(%d).Uint32()`, tag)
		case model.KindUint64:
			w.writef(`return m.msg.Field(%d).Uint64()`, tag)

		case model.KindBin64:
			w.writef(`return m.msg.Field(%d).Bin64()`, tag)
		case model.KindBin128:
			w.writef(`return m.msg.Field(%d).Bin128()`, tag)
		case model.KindBin256:
			w.writef(`return m.msg.Field(%d).Bin256()`, tag)

		case model.KindFloat32:
			w.writef(`return m.msg.Field(%d).Float32()`, tag)
		case model.KindFloat64:
			w.writef(`return m.msg.Field(%d).Float64()`, tag)

		case model.KindBytes:
			w.writef(`return m.msg.Field(%d).Bytes()`, tag)
		case model.KindString:
			w.writef(`return m.msg.Field(%d).String()`, tag)

		case model.KindAny:
			w.writef(`return m.msg.Field(%d)`, tag)
		case model.KindAnyMessage:
			w.writef(`return m.msg.Field(%d).Message()`, tag)
		}

		w.writef(`}`)
		w.line()

	case model.KindList:
		decodeFunc := typeDecodeRefFunc(field.Type.Element)

		w.writef(`func (m %v) %v() %v {`, def.Name, fieldName, typeName)
		w.writef(`return spec.NewTypedList(m.msg.FieldBytes(%d), %v)`, tag, decodeFunc)
		w.writef(`}`)
		w.line()

	case model.KindEnum,
		model.KindMessage,
		model.KindStruct:
		newFunc := typeNewFunc(field.Type)

		w.writef(`func (m %v) %v() %v {`, def.Name, fieldName, typeName)
		w.writef(`return %v(m.msg.FieldBytes(%d))`, newFunc, tag)
		w.writef(`}`)
		w.line()
	}
	return nil
}

func (w *writer) messageHasFields(def *model.Definition) error {
	fields := def.Message.Fields.List

	for _, field := range fields {
		if err := w.messageHasField(def, field); err != nil {
			return err
		}
	}

	w.line()
	return nil
}

func (w *writer) messageHasField(def *model.Definition, field *model.Field) error {
	fieldName := messageFieldName(field)
	tag := field.Tag

	w.writef(`func (m %v) Has%v() bool {`, def.Name, fieldName)
	w.writef(`return m.msg.HasField(%d)`, tag)
	w.writef(`}`)
	w.line()
	return nil
}

func (w *writer) messageMethods(def *model.Definition) error {
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

// writer

func (w *writer) messageWriter(def *model.Definition) error {
	if err := w.messageWriterDef(def); err != nil {
		return err
	}
	if err := w.newMessageWriter(def); err != nil {
		return err
	}
	if err := w.messageWriterFields(def); err != nil {
		return err
	}
	if err := w.messageWriterBuild(def); err != nil {
		return err
	}
	return nil
}

func (w *writer) messageWriterDef(def *model.Definition) error {
	w.linef(`// %vWriter`, def.Name)
	w.line()
	w.linef(`type %vWriter struct {`, def.Name)
	w.line(`w spec.MessageWriter`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) newMessageWriter(def *model.Definition) error {
	w.linef(`func New%vWriter() %vWriter {`, def.Name, def.Name)
	w.linef(`w := spec.NewMessageWriter()`)
	w.linef(`return %vWriter{w}`, def.Name)
	w.linef(`}`)
	w.line()

	w.linef(`func New%vWriterBuffer(b buffer.Buffer) %vWriter {`, def.Name, def.Name)
	w.linef(`w := spec.NewMessageWriterBuffer(b)`)
	w.linef(`return %vWriter{w}`, def.Name)
	w.linef(`}`)
	w.line()

	w.linef(`func New%vWriterTo(w spec.MessageWriter) %vWriter {`, def.Name, def.Name)
	w.linef(`return %vWriter{w}`, def.Name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) messageWriterBuild(def *model.Definition) error {
	w.linef(`func (w %vWriter) End() error {`, def.Name)
	w.linef(`return w.w.End()`)
	w.linef(`}`)
	w.line()

	w.linef(`func (w %vWriter) Build() (_ %v, err error) {`, def.Name, def.Name)
	w.linef(`bytes, err := w.w.Build()`)
	w.linef(`if err != nil {
		return
	}`)
	w.linef(`return New%v(bytes), nil`, def.Name)
	w.linef(`}`)
	w.line()

	w.linef(`func (w %vWriter) Unwrap() spec.MessageWriter {`, def.Name)
	w.linef(`return w.w`)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) messageWriterFields(def *model.Definition) error {
	fields := def.Message.Fields.List

	for _, field := range fields {
		if err := w.messageWriterField(def, field); err != nil {
			return err
		}
	}

	w.line()
	return nil
}

func (w *writer) messageWriterField(def *model.Definition, field *model.Field) error {
	fname := messageFieldName(field)
	tname := inTypeName(field.Type)
	wname := fmt.Sprintf("%vWriter", def.Name)

	tag := field.Tag
	kind := field.Type.Kind

	switch kind {
	default:
		w.linef(`func (w %vWriter) %v(v %v) %v {`, def.Name, fname, tname, wname)

		switch kind {
		case model.KindBool:
			w.linef(`w.w.Field(%d).Bool(v)`, tag)
		case model.KindByte:
			w.linef(`w.w.Field(%d).Byte(v)`, tag)

		case model.KindInt16:
			w.linef(`w.w.Field(%d).Int16(v)`, tag)
		case model.KindInt32:
			w.linef(`w.w.Field(%d).Int32(v)`, tag)
		case model.KindInt64:
			w.linef(`w.w.Field(%d).Int64(v)`, tag)

		case model.KindUint16:
			w.linef(`w.w.Field(%d).Uint16(v)`, tag)
		case model.KindUint32:
			w.linef(`w.w.Field(%d).Uint32(v)`, tag)
		case model.KindUint64:
			w.linef(`w.w.Field(%d).Uint64(v)`, tag)

		case model.KindBin64:
			w.linef(`w.w.Field(%d).Bin64(v)`, tag)
		case model.KindBin128:
			w.linef(`w.w.Field(%d).Bin128(v)`, tag)
		case model.KindBin256:
			w.linef(`w.w.Field(%d).Bin256(v)`, tag)

		case model.KindFloat32:
			w.linef(`w.w.Field(%d).Float32(v)`, tag)
		case model.KindFloat64:
			w.linef(`w.w.Field(%d).Float64(v)`, tag)

		case model.KindBytes:
			w.linef(`w.w.Field(%d).Bytes(v)`, tag)
		case model.KindString:
			w.linef(`w.w.Field(%d).String(v)`, tag)
		}

		w.linef(`return w`)
		w.linef(`}`)
		w.line()

	case model.KindAny:
		w.linef(`func (w %v) %v() spec.FieldWriter {`, wname, fname)
		w.linef(`return w.w.Field(%d)`, tag)
		w.linef(`}`)
		w.line()

		w.linef(`func (w %v) Copy%v(v spec.Value) error {`, wname, fname)
		w.linef(`return w.w.Field(%d).Any(v)`, tag)
		w.linef(`}`)
		w.line()

	case model.KindAnyMessage:
		w.linef(`func (w %v) %v() spec.MessageWriter {`, wname, fname)
		w.linef(`return w.w.Field(%d).Message()`, tag)
		w.linef(`}`)
		w.line()

		w.linef(`func (w %v) Copy%v(v spec.Message) error {`, wname, fname)
		w.linef(`return w.w.Field(%d).Any(v.Raw())`, tag)
		w.linef(`}`)
		w.line()

	case model.KindEnum:
		writeFunc := typeWriteFunc(field.Type)

		w.linef(`func (w %v) %v(v %v) %v {`, wname, fname, tname, wname)
		w.linef(`spec.WriteField(w.w.Field(%d), v, %v)`, tag, writeFunc)
		w.linef(`return w`)
		w.linef(`}`)
		w.line()

	case model.KindStruct:
		writeFunc := typeWriteFunc(field.Type)

		w.linef(`func (w %v) %v(v %v) %v {`, wname, fname, tname, wname)
		w.linef(`spec.WriteField(w.w.Field(%d), v, %v)`, tag, writeFunc)
		w.linef(`return w`)
		w.linef(`}`)
		w.line()

	case model.KindList:
		writer := typeWriter(field.Type)
		buildList := typeWriteFunc(field.Type)
		encodeElement := typeWriteFunc(field.Type.Element)

		w.linef(`func (w %v) %v() %v {`, wname, fname, writer)
		w.linef(`w1 := w.w.Field(%d).List()`, tag)
		w.linef(`return %v(w1, %v)`, buildList, encodeElement)
		w.linef(`}`)
		w.line()

	case model.KindMessage:
		writer := typeWriter(field.Type)
		newMessageWriter := typeWriteFunc(field.Type)
		w.linef(`func (w %v) %v() %v {`, wname, fname, writer)
		w.linef(`w1 := w.w.Field(%d).Message()`, tag)
		w.linef(`return %v(w1)`, newMessageWriter)
		w.linef(`}`)
		w.line()

		tname := typeName(field.Type)
		w.linef(`func (w %v) Copy%v(v %v) error {`, wname, fname, tname)
		w.linef(`return w.w.Field(%d).Any(v.Unwrap().Raw())`, tag)
		w.linef(`}`)
		w.line()
	}
	return nil
}

// util

func messageFieldName(field *model.Field) string {
	return toUpperCamelCase(field.Name)
}
