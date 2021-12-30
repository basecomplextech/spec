package golang

import (
	"fmt"

	"github.com/baseone-run/spec/compiler"
)

func (w *writer) struct_(def *compiler.Definition) error {
	if err := w.structDef(def); err != nil {
		return err
	}
	if err := w.structReadFunc(def); err != nil {
		return err
	}
	if err := w.structRead(def); err != nil {
		return err
	}
	if err := w.structWrite(def); err != nil {
		return err
	}
	return nil
}

func (w *writer) structDef(def *compiler.Definition) error {
	w.linef("type %v struct {", def.Name)
	w.line("}")
	w.line()
	return nil
}

func (w *writer) structReadFunc(def *compiler.Definition) error {
	w.linef(`func Read%v(b []byte) (str %v, err error) {`, def.Name, def.Name)
	w.line(`if len(b) == 0 {
		return
	}`)
	w.line(`err = str.Read(b)`)
	w.line(`return`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) structRead(def *compiler.Definition) error {
	w.linef(`func (m *%v) Read(b []byte) error {`, def.Name)
	w.line(`return nil`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) structWrite(def *compiler.Definition) error {
	w.linef(`func (m %v) Write(w spec.Writer) error {`, def.Name)
	w.line(`return nil`)
	w.line("}")
	return nil
}

func structFieldName(field *compiler.StructField) string {
	return toUpperCamelCase(field.Name)
}

func structReadFunc(typ *compiler.Type) string {
	if typ.Kind != compiler.KindStruct {
		panic(fmt.Sprintf("must be struct, got=%v", typ.Kind))
	}

	if typ.Import == nil {
		return fmt.Sprintf("Read%v", typ.Name)
	}
	return fmt.Sprintf("%v.Read%v", typ.ImportName, typ.Name)
}
