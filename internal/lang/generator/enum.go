// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package generator

import (
	"fmt"
	"strings"

	"github.com/basecomplextech/spec/internal/lang/model"
)

type enumWriter struct {
	*writer
}

func newEnumWriter(w *writer) *enumWriter {
	return &enumWriter{w}
}

func (w *enumWriter) enum(def *model.Definition) error {
	if err := w.def(def); err != nil {
		return err
	}
	if err := w.values(def); err != nil {
		return err
	}
	if err := w.open_method(def); err != nil {
		return err
	}
	if err := w.decode_method(def); err != nil {
		return err
	}
	if err := w.encode_method(def); err != nil {
		return err
	}
	if err := w.string_method(def); err != nil {
		return err
	}
	return nil
}

func (w *enumWriter) def(def *model.Definition) error {
	w.linef(`// %v`, def.Name)
	w.line()
	w.linef("type %v int32", def.Name)
	w.line()
	return nil
}

func (w *enumWriter) values(def *model.Definition) error {
	w.line("const (")

	for _, val := range def.Enum.Values {
		// EnumValue Enum = 1
		name := enumValueName(val)
		w.linef("%v %v = %d", name, def.Name, val.Number)
	}

	w.line(")")
	w.line()
	return nil
}

func (w *enumWriter) open_method(def *model.Definition) error {
	name := def.Name
	w.linef(`func Open%v(b []byte) %v {`, name, name)
	w.linef(`v, _, _ := spec.DecodeInt32(b)`)
	w.linef(`return %v(v)`, name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *enumWriter) decode_method(def *model.Definition) error {
	name := def.Name
	w.linef(`func Decode%v(b []byte) (result %v, size int, err error) {`, name, name)
	w.linef(`v, size, err := spec.DecodeInt32(b)`)
	w.linef(`if err != nil || size == 0 {
		return
	}`)
	w.linef(`result = %v(v)`, name)
	w.line(`return`)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *enumWriter) encode_method(def *model.Definition) error {
	w.linef(`func Encode%vTo(b buffer.Buffer, v %v) (int, error) {`, def.Name, def.Name)
	w.linef(`return spec.EncodeInt32(b, int32(v))`)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *enumWriter) string_method(def *model.Definition) error {
	w.linef("func (e %v) String() string {", def.Name)
	w.line("switch e {")

	for _, val := range def.Enum.Values {
		name := enumValueName(val)
		w.linef("case %v:", name)
		w.linef(`return "%v"`, strings.ToLower(val.Name))
	}

	w.line("}")
	w.line(`return ""`)
	w.line("}")
	w.line()
	return nil
}

func enumValueName(val *model.EnumValue) string {
	name := toUpperCamelCase(val.Name)
	return fmt.Sprintf("%v_%v", val.Enum.Def.Name, name)
}
