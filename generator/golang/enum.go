package golang

import (
	"strings"

	"github.com/baseone-run/spec/compiler"
)

func (w *writer) enum(def *compiler.Definition) error {
	if err := w.enumDef(def); err != nil {
		return err
	}
	if err := w.enumValues(def); err != nil {
		return err
	}
	if err := w.enumString(def); err != nil {
		return err
	}
	return nil
}

func (w *writer) enumDef(def *compiler.Definition) error {
	w.linef("type %v int32", def.Name)
	w.line()
	return nil
}

func (w *writer) enumValues(def *compiler.Definition) error {
	w.line("const (")

	for _, val := range def.Enum.Values {
		// EnumValue Enum = 1
		name := goEnumValueName(val)
		w.linef("%v %v = %d", name, def.Name, val.Number)
	}

	w.line(")")
	w.line()
	return nil
}

func (w *writer) enumString(def *compiler.Definition) error {
	w.linef("func (e %v) String() string {", def.Name)
	w.line("switch e {")

	for _, val := range def.Enum.Values {
		name := goEnumValueName(val)
		w.linef("case %v:", name)
		w.linef(`return "%v"`, strings.ToLower(val.Name))
	}

	w.line("}")
	w.line(`return ""`)
	w.line("}")
	w.line()
	return nil
}

func goEnumValueName(val *compiler.EnumValue) string {
	name := toUpperCamelCase(val.Name)
	return val.Enum.Def.Name + name
}
