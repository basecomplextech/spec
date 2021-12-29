package golang

import "github.com/baseone-run/spec/compiler"

func (w *writer) struct_(def *compiler.Definition) error {
	return nil
}

func (w *writer) readStruct(def *compiler.Definition) error {
	return nil
}

func (w *writer) writeStruct(def *compiler.Definition) error {
	return nil
}

func goStructFieldName(field *compiler.StructField) string {
	return toUpperCamelCase(field.Name)
}
