// Copyright 2022 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package generator

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/model"
)

type messageWriter struct {
	*writer
}

func newMessageWriter(w *writer) *messageWriter {
	return &messageWriter{w}
}

func (w *messageWriter) message(def *model.Definition) error {
	if err := w.def(def); err != nil {
		return err
	}
	if err := w.new_methods(def); err != nil {
		return err
	}
	if err := w.parse_method(def); err != nil {
		return err
	}
	if err := w.fields(def); err != nil {
		return err
	}
	if err := w.has_fields(def); err != nil {
		return err
	}
	if err := w.methods(def); err != nil {
		return err
	}
	return nil
}

func (w *messageWriter) def(def *model.Definition) error {
	w.linef(`// %v`, def.Name)
	w.line()
	w.linef(`type %v struct {`, def.Name)
	w.line(`msg spec.Message`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *messageWriter) new_methods(def *model.Definition) error {
	w.linef(`func New%v(msg spec.Message) %v {`, def.Name, def.Name)
	w.linef(`return %v{msg}`, def.Name)
	w.linef(`}`)
	w.line()

	w.linef(`func Open%v(b []byte) %v {`, def.Name, def.Name)
	w.linef(`msg := spec.OpenMessage(b)`)
	w.linef(`return %v{msg}`, def.Name)
	w.linef(`}`)
	w.line()

	w.linef(`func Open%vErr(b []byte) (_ %v, err error) {`, def.Name, def.Name)
	w.linef(`msg, err := spec.OpenMessageErr(b)`)
	w.line(`if err != nil {
		return
	}`)
	w.linef(`return %v{msg}, nil`, def.Name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *messageWriter) parse_method(def *model.Definition) error {
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

func (w *messageWriter) fields(def *model.Definition) error {
	fields := def.Message.Fields.List

	for _, field := range fields {
		if err := w.field(def, field); err != nil {
			return err
		}
	}

	if len(fields) > 1 {
		w.line()
	}
	return nil
}

func (w *messageWriter) field(def *model.Definition, field *model.Field) error {
	fieldName := messageFieldName(field)
	typeName := typeRefName(field.Type)

	tag := field.Tag
	kind := field.Type.Kind

	switch kind {
	default:
		w.writef(`func (m %v) %v() %v {`, def.Name, fieldName, typeName)

		switch kind {
		case model.KindBool:
			w.writef(`return m.msg.Bool(%d)`, tag)
		case model.KindByte:
			w.writef(`return m.msg.Byte(%d)`, tag)

		case model.KindInt16:
			w.writef(`return m.msg.Int16(%d)`, tag)
		case model.KindInt32:
			w.writef(`return m.msg.Int32(%d)`, tag)
		case model.KindInt64:
			w.writef(`return m.msg.Int64(%d)`, tag)

		case model.KindUint16:
			w.writef(`return m.msg.Uint16(%d)`, tag)
		case model.KindUint32:
			w.writef(`return m.msg.Uint32(%d)`, tag)
		case model.KindUint64:
			w.writef(`return m.msg.Uint64(%d)`, tag)

		case model.KindBin64:
			w.writef(`return m.msg.Bin64(%d)`, tag)
		case model.KindBin128:
			w.writef(`return m.msg.Bin128(%d)`, tag)
		case model.KindBin256:
			w.writef(`return m.msg.Bin256(%d)`, tag)

		case model.KindFloat32:
			w.writef(`return m.msg.Float32(%d)`, tag)
		case model.KindFloat64:
			w.writef(`return m.msg.Float64(%d)`, tag)

		case model.KindBytes:
			w.writef(`return m.msg.Bytes(%d)`, tag)
		case model.KindString:
			w.writef(`return m.msg.String(%d)`, tag)

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
		w.writef(`return spec.OpenTypedList(m.msg.FieldRaw(%d), %v)`, tag, decodeFunc)
		w.writef(`}`)
		w.line()

	case model.KindMessage:
		makeFunc := typeMakeMessageFunc(field.Type)

		w.writef(`func (m %v) %v() %v {`, def.Name, fieldName, typeName)
		w.writef(`return %v(m.msg.Message(%d))`, makeFunc, tag)
		w.writef(`}`)
		w.line()

	case model.KindEnum,
		model.KindStruct:
		newFunc := typeNewFunc(field.Type)

		w.writef(`func (m %v) %v() %v {`, def.Name, fieldName, typeName)
		w.writef(`return %v(m.msg.FieldRaw(%d))`, newFunc, tag)
		w.writef(`}`)
		w.line()
	}
	return nil
}

func (w *messageWriter) has_fields(def *model.Definition) error {
	fields := def.Message.Fields.List

	for _, field := range fields {
		if err := w.has_field(def, field); err != nil {
			return err
		}
	}

	if len(fields) > 1 {
		w.line()
	}
	return nil
}

func (w *messageWriter) has_field(def *model.Definition, field *model.Field) error {
	fieldName := messageFieldName(field)
	tag := field.Tag

	w.writef(`func (m %v) Has%v() bool {`, def.Name, fieldName)
	w.writef(`return m.msg.HasField(%d)`, tag)
	w.writef(`}`)
	w.line()
	return nil
}

func (w *messageWriter) methods(def *model.Definition) error {
	w.writef(`func (m %v) IsEmpty() bool {`, def.Name)
	w.writef(`return m.msg.Empty()`)
	w.writef(`}`)
	w.line()

	w.writef(`func (m %v) Clone() %v {`, def.Name, def.Name)
	w.writef(`return %v{m.msg.Clone()}`, def.Name)
	w.writef(`}`)
	w.line()

	w.writef(`func (m %v) CloneToArena(a alloc.Arena) %v {`, def.Name, def.Name)
	w.writef(`return %v{m.msg.CloneToArena(a)}`, def.Name)
	w.writef(`}`)
	w.line()

	w.writef(`func (m %v) CloneToBuffer(b buffer.Buffer) %v {`, def.Name, def.Name)
	w.writef(`return %v{m.msg.CloneToBuffer(b)}`, def.Name)
	w.writef(`}`)
	w.line()

	w.writef(`func (m %v) Unwrap() spec.Message {`, def.Name)
	w.writef(`return m.msg`)
	w.writef(`}`)
	w.line()
	return nil
}

// writer

func (w *messageWriter) messageWriter(def *model.Definition) error {
	if err := w.writer_def(def); err != nil {
		return err
	}
	if err := w.writer_new_method(def); err != nil {
		return err
	}
	if err := w.writer_fields(def); err != nil {
		return err
	}
	if err := w.writer_end(def); err != nil {
		return err
	}
	return nil
}

func (w *messageWriter) writer_def(def *model.Definition) error {
	w.linef(`// %vWriter`, def.Name)
	w.line()
	w.linef(`type %vWriter struct {`, def.Name)
	w.line(`w spec.MessageWriter`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *messageWriter) writer_new_method(def *model.Definition) error {
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

func (w *messageWriter) writer_end(def *model.Definition) error {
	w.linef(`func (w %vWriter) Merge(msg %v) error {`, def.Name, def.Name)
	w.linef(`return w.w.Merge(msg.Unwrap())`)
	w.linef(`}`)
	w.line()

	w.linef(`func (w %vWriter) End() error {`, def.Name)
	w.linef(`return w.w.End()`)
	w.linef(`}`)
	w.line()

	w.linef(`func (w %vWriter) Build() (_ %v, err error) {`, def.Name, def.Name)
	w.linef(`bytes, err := w.w.Build()`)
	w.linef(`if err != nil {
		return
	}`)
	w.linef(`return Open%vErr(bytes)`, def.Name)
	w.linef(`}`)
	w.line()

	w.linef(`func (w %vWriter) Unwrap() spec.MessageWriter {`, def.Name)
	w.linef(`return w.w`)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *messageWriter) writer_fields(def *model.Definition) error {
	fields := def.Message.Fields.List

	for _, field := range fields {
		if err := w.writer_field(def, field); err != nil {
			return err
		}
	}

	w.line()
	return nil
}

func (w *messageWriter) writer_field(def *model.Definition, field *model.Field) error {
	fname := messageFieldName(field)
	tname := inTypeName(field.Type)
	wname := fmt.Sprintf("%vWriter", def.Name)

	tag := field.Tag
	kind := field.Type.Kind

	switch kind {
	default:
		w.writef(`func (w %vWriter) %v(v %v) {`, def.Name, fname, tname)

		switch kind {
		case model.KindBool:
			w.writef(`w.w.Field(%d).Bool(v)`, tag)
		case model.KindByte:
			w.writef(`w.w.Field(%d).Byte(v)`, tag)

		case model.KindInt16:
			w.writef(`w.w.Field(%d).Int16(v)`, tag)
		case model.KindInt32:
			w.writef(`w.w.Field(%d).Int32(v)`, tag)
		case model.KindInt64:
			w.writef(`w.w.Field(%d).Int64(v)`, tag)

		case model.KindUint16:
			w.writef(`w.w.Field(%d).Uint16(v)`, tag)
		case model.KindUint32:
			w.writef(`w.w.Field(%d).Uint32(v)`, tag)
		case model.KindUint64:
			w.writef(`w.w.Field(%d).Uint64(v)`, tag)

		case model.KindBin64:
			w.writef(`w.w.Field(%d).Bin64(v)`, tag)
		case model.KindBin128:
			w.writef(`w.w.Field(%d).Bin128(v)`, tag)
		case model.KindBin256:
			w.writef(`w.w.Field(%d).Bin256(v)`, tag)

		case model.KindFloat32:
			w.writef(`w.w.Field(%d).Float32(v)`, tag)
		case model.KindFloat64:
			w.writef(`w.w.Field(%d).Float64(v)`, tag)

		case model.KindBytes:
			w.writef(`w.w.Field(%d).Bytes(v)`, tag)
		case model.KindString:
			w.writef(`w.w.Field(%d).String(v)`, tag)
		}
		w.linef(`}`)

	case model.KindAny:
		w.writef(`func (w %v) %v() spec.FieldWriter {`, wname, fname)
		w.writef(`return w.w.Field(%d)`, tag)
		w.linef(`}`)

		w.writef(`func (w %v) Copy%v(v spec.Value) error {`, wname, fname)
		w.writef(`return w.w.Field(%d).Any(v)`, tag)
		w.linef(`}`)

	case model.KindAnyMessage:
		w.writef(`func (w %v) %v() spec.MessageWriter {`, wname, fname)
		w.writef(`return w.w.Field(%d).Message()`, tag)
		w.linef(`}`)

		w.writef(`func (w %v) Copy%v(v spec.Message) error {`, wname, fname)
		w.writef(`return w.w.Field(%d).Any(v.Raw())`, tag)
		w.linef(`}`)

	case model.KindEnum:
		writeFunc := typeWriteFunc(field.Type)

		w.writef(`func (w %v) %v(v %v) {`, wname, fname, tname)
		w.writef(`spec.WriteField(w.w.Field(%d), v, %v)`, tag, writeFunc)
		w.linef(`}`)

	case model.KindStruct:
		writeFunc := typeWriteFunc(field.Type)

		w.writef(`func (w %v) %v(v %v) {`, wname, fname, tname)
		w.writef(`spec.WriteField(w.w.Field(%d), v, %v)`, tag, writeFunc)
		w.linef(`}`)

	case model.KindList:
		writer := typeWriter(field.Type)
		buildList := typeWriteFunc(field.Type)
		encodeElement := typeWriteFunc(field.Type.Element)

		w.linef(`func (w %v) %v() %v {`, wname, fname, writer)
		w.linef(`w1 := w.w.Field(%d).List()`, tag)
		w.linef(`return %v(w1, %v)`, buildList, encodeElement)
		w.linef(`}`)

	case model.KindMessage:
		writer := typeWriter(field.Type)
		writer_new_method := typeWriteFunc(field.Type)
		w.linef(`func (w %v) %v() %v {`, wname, fname, writer)
		w.linef(`w1 := w.w.Field(%d).Message()`, tag)
		w.linef(`return %v(w1)`, writer_new_method)
		w.linef(`}`)

		tname := typeName(field.Type)
		w.linef(`func (w %v) Copy%v(v %v) error {`, wname, fname, tname)
		w.linef(`return w.w.Field(%d).Any(v.Unwrap().Raw())`, tag)
		w.linef(`}`)
	}
	return nil
}

// util

func messageFieldName(field *model.Field) string {
	return toUpperCamelCase(field.Name)
}
