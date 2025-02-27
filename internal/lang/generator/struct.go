// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package generator

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/model"
)

type structWriter struct {
	*writer
}

func newStructWriter(w *writer) *structWriter {
	return &structWriter{w}
}

func (w *structWriter) struct_(def *model.Definition) error {
	if err := w.def(def); err != nil {
		return err
	}
	if err := w.open(def); err != nil {
		return err
	}
	if err := w.decode_method(def); err != nil {
		return err
	}
	if err := w.encode_method(def); err != nil {
		return err
	}
	return nil
}

func (w *structWriter) def(def *model.Definition) error {
	w.linef(`// %v`, def.Name)
	w.line()
	w.linef("type %v struct {", def.Name)

	fields := def.Struct.Fields.Values()
	for _, field := range fields {
		name := structFieldName(field)
		typ := typeName(field.Type)
		goTag := fmt.Sprintf("`json:\"%v\"`", field.Name)
		w.linef("%v %v %v", name, typ, goTag)
	}

	w.line("}")
	w.line()
	return nil
}

func (w *structWriter) open(def *model.Definition) error {
	w.linef(`func Open%v(b []byte) %v {`, def.Name, def.Name)
	w.linef(`s, _, _ := Decode%v(b)`, def.Name)
	w.line(`return s`)
	w.line(`}`)
	w.line()

	w.linef(`func Decode%v(b []byte) (s %v, size int, err error) {`, def.Name, def.Name)
	w.line(`size, err = s.Decode(b)`)
	w.line(`return s, size, err`)
	w.line(`}`)
	w.line()

	w.linef(`func Encode%vTo(b buffer.Buffer, s %v) (int, error) {`, def.Name, def.Name)
	w.line(`return s.EncodeTo(b)`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *structWriter) decode_method(def *model.Definition) error {
	w.linef(`func (s *%v) Decode(b []byte) (size int, err error) {`, def.Name)
	w.line(`dataSize, size, err := spec.DecodeStruct(b)`)
	w.line(`if err != nil || size == 0 {
		return
	}`)
	w.line()

	w.line(`b = b[len(b)-size:]
	n := size - dataSize
	off := len(b) - n
	`)
	w.line()

	w.line(`// Decode in reverse order`)
	w.line()

	fields := def.Struct.Fields.Values()
	for i := len(fields) - 1; i >= 0; i-- {
		field := fields[i]
		fieldName := structFieldName(field)
		decodeName := typeDecodeFunc(field.Type)
		if field.Type.Kind == model.KindString {
			decodeName = "spec.DecodeStringClone"
		}

		w.linef(`s.%v, n, err = %v(b[:off])`, fieldName, decodeName)
		w.line(`if err != nil {
			return
		}`)
		w.line(`off -= n`)
		w.line()
	}

	w.line(`return size, err`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *structWriter) encode_method(def *model.Definition) error {
	w.linef(`func (s %v) EncodeTo(b buffer.Buffer) (int, error) {`, def.Name)
	w.line(`var dataSize, n int`)
	w.line(`var err error`)
	w.line()

	fields := def.Struct.Fields.Values()
	for _, field := range fields {
		fieldName := structFieldName(field)
		writeFunc := typeWriteFunc(field.Type)

		w.linef(`n, err = %v(b, s.%v)`, writeFunc, fieldName)
		w.line(`if err != nil {
			return 0, err
		}`)
		w.line(`dataSize += n`)
		w.line()
	}

	w.line(`n, err = spec.EncodeStruct(b, dataSize)`)
	w.line(`if err != nil {
			return 0, err
		}`)
	w.line(`return dataSize + n, nil`)
	w.line(`}`)
	w.line()
	return nil
}

func structFieldName(field *model.StructField) string {
	return toUpperCamelCase(field.Name)
}
