package golang

import (
	"fmt"
	"strings"

	"github.com/complexl/spec/compiler"
)

func (w *writer) enum(def *compiler.Definition) error {
	if err := w.enumDef(def); err != nil {
		return err
	}
	if err := w.enumValues(def); err != nil {
		return err
	}
	if err := w.readEnum(def); err != nil {
		return err
	}
	if err := w.enumWrite(def); err != nil {
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

func (w *writer) readEnum(def *compiler.Definition) error {
	name := def.Name
	w.linef(`func Read%v(b []byte) (%v, error) {`, name, name)
	w.linef(`r := spec.NewReader(b)`)
	w.linef(`v, _, err := r.ReadInt32()`)
	w.linef(`if err != nil {
		return 0, err
	}`)
	w.linef(`return %v(v), nil`, name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) enumWrite(def *compiler.Definition) error {
	w.linef(`func (e %v) Write(w *spec.Writer) error {`, def.Name)
	w.linef(`return w.Int32(int32(e))`)
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

func enumReadFunc(typ *compiler.Type) string {
	if typ.Import == nil {
		return fmt.Sprintf("Read%v", typ.Name)
	}
	return fmt.Sprintf("%v.Read%v", typ.ImportName, typ.Name)
}
