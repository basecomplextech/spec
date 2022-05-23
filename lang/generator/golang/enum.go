package golang

import (
	"strings"

	"github.com/epochtimeout/spec/lang/compiler"
)

func (w *writer) enum(def *compiler.Definition) error {
	if err := w.enumDef(def); err != nil {
		return err
	}
	if err := w.enumValues(def); err != nil {
		return err
	}
	if err := w.getEnum(def); err != nil {
		return err
	}
	if err := w.decodeEnum(def); err != nil {
		return err
	}
	if err := w.encodeEnum(def); err != nil {
		return err
	}
	if err := w.enumString(def); err != nil {
		return err
	}
	return nil
}

func (w *writer) enumDef(def *compiler.Definition) error {
	w.linef(`// %v`, def.Name)
	w.line()
	w.linef("type %v int32", def.Name)
	w.line()
	return nil
}

func (w *writer) enumValues(def *compiler.Definition) error {
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

func (w *writer) getEnum(def *compiler.Definition) error {
	name := def.Name
	w.linef(`func Get%v(b []byte) %v {`, name, name)
	w.linef(`v, _, _ := spec.DecodeInt32(b)`)
	w.linef(`return %v(v)`, name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) decodeEnum(def *compiler.Definition) error {
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

func (w *writer) encodeEnum(def *compiler.Definition) error {
	w.linef(`func Encode%v(b buffer.Buffer, v %v) (int, error) {`, def.Name, def.Name)
	w.linef(`return spec.EncodeInt32(b, int32(v))`)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) enumString(def *compiler.Definition) error {
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

func enumValueName(val *compiler.EnumValue) string {
	name := toUpperCamelCase(val.Name)
	return val.Enum.Def.Name + name
}
