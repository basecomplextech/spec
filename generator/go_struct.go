package generator

import "github.com/baseone-run/spec/compiler"

func (w *goWriter) struct_(def *compiler.Definition) error {
	return nil
}

func (w *goWriter) readStruct(def *compiler.Definition) error {
	return nil
}

func (w *goWriter) writeStruct(def *compiler.Definition) error {
	return nil
}

func goStructFieldName(field *compiler.StructField) string {
	return toUpperCamelCase(field.Name)
}
