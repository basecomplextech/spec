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
	w.linef(`func Encode%v(e *spec.Encoder, s %v) ([]byte, error) {`, def.Name, def.Name)
	w.linef(`if err := e.BeginStruct(); err != nil {
		return nil, err
	}`)
	w.line()

	for _, field := range def.Struct.Fields {
		fieldName := structFieldName(field)
		kind := field.Type.Kind

		switch kind {
		case compiler.KindBool:
			w.linef(`e.Bool(s.%v)`, fieldName)
		case compiler.KindByte:
			w.linef(`e.Byte(s.%v)`, fieldName)

		case compiler.KindInt32:
			w.linef(`e.Int32(s.%v)`, fieldName)
		case compiler.KindInt64:
			w.linef(`e.Int64(s.%v)`, fieldName)
		case compiler.KindUint32:
			w.linef(`e.Uint32(s.%v)`, fieldName)
		case compiler.KindUint64:
			w.linef(`e.Uint64(s.%v)`, fieldName)

		case compiler.KindU128:
			w.linef(`e.U128(s.%v)`, fieldName)
		case compiler.KindU256:
			w.linef(`e.U256(s.%v)`, fieldName)

		case compiler.KindFloat32:
			w.linef(`e.Float32(s.%v)`, fieldName)
		case compiler.KindFloat64:
			w.linef(`e.Float64(s.%v)`, fieldName)

		case compiler.KindBytes:
			w.linef(`e.Bytes(s.%v)`, fieldName)
		case compiler.KindString:
			w.linef(`e.String(s.%v)`, fieldName)

		case compiler.KindStruct:
			decodeFunc := typeDecodeFunc(field.Type)
			w.linef(`%v(e, s.%v)`, decodeFunc, fieldName)
		}

		w.linef(`e.StructField()`)
		w.line()
	}

	// end
	w.line(`return e.End()`)
	w.line(`}`)
	return nil
}

func structFieldName(field *compiler.StructField) string {
	return toUpperCamelCase(field.Name)
}
