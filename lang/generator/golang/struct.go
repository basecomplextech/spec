package golang

import (
	"fmt"

	"github.com/complexl/spec/lang/compiler"
)

func (w *writer) struct_(def *compiler.Definition) error {
	if err := w.structDef(def); err != nil {
		return err
	}
	if err := w.getStruct(def); err != nil {
		return err
	}
	if err := w.decodeStruct(def); err != nil {
		return err
	}
	if err := w.encodeStruct(def); err != nil {
		return err
	}
	return nil
}

func (w *writer) structDef(def *compiler.Definition) error {
	w.linef(`// %v`, def.Name)
	w.line()
	w.linef("type %v struct {", def.Name)

	for _, field := range def.Struct.Fields {
		name := structFieldName(field)
		typ := typeName(field.Type)
		goTag := fmt.Sprintf("`json:\"%v\"`", field.Name)
		w.linef("%v %v %v", name, typ, goTag)
	}

	w.line("}")
	w.line()
	return nil
}

func (w *writer) getStruct(def *compiler.Definition) error {
	w.linef(`func Get%v(b []byte) (result %v) {`, def.Name, def.Name)
	w.linef(`result, _, _ = Decode%v(b)`, def.Name)
	w.line(`return`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) decodeStruct(def *compiler.Definition) error {
	w.linef(`func Decode%v(b []byte) (result %v, size int, err error) {`, def.Name, def.Name)
	w.line(`dataSize, size, err := spec.DecodeStruct(b)`)
	w.line(`if err != nil {
		return
	}`)
	w.line()

	w.line(`b = b[len(b)-size:]
	n := size - dataSize
	off := len(b)
	`)
	w.line()

	w.line(`// decode in reverse order`)
	w.line()

	fields := def.Struct.Fields
	for i := len(fields) - 1; i >= 0; i-- {
		field := fields[i]
		fieldName := structFieldName(field)
		decodeName := typeDecodeFunc(field.Type)

		w.line(`off -= n`)
		w.linef(`result.%v, n, err = %v(b[:off])`, fieldName, decodeName)
		w.line(`if err != nil {
			return
		}`)
		w.line()
	}

	w.line(`return`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) encodeStruct(def *compiler.Definition) error {
	w.linef(`func Encode%v(b spec.Buffer, s %v) (int, error) {`, def.Name, def.Name)
	w.line(`var dataSize, n int`)
	w.line(`var err error`)
	w.line()

	for _, field := range def.Struct.Fields {
		fieldName := structFieldName(field)
		encodeFunc := typeEncodeFunc(field.Type)

		w.linef(`n, err = %v(b, s.%v)`, encodeFunc, fieldName)
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

func structFieldName(field *compiler.StructField) string {
	return toUpperCamelCase(field.Name)
}
