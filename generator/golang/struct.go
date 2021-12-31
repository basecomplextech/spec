package golang

import (
	"fmt"

	"github.com/baseone-run/spec/compiler"
)

func (w *writer) struct_(def *compiler.Definition) error {
	if err := w.structDef(def); err != nil {
		return err
	}
	if err := w.readStruct(def); err != nil {
		return err
	}
	if err := w.structMarshal(def); err != nil {
		return err
	}
	if err := w.structUnmarshal(def); err != nil {
		return err
	}
	if err := w.structWrite(def); err != nil {
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
		typ := objectType(field.Type)
		tag := fmt.Sprintf("`json:\"%v\"`", field.Name)
		w.linef("%v %v %v", name, typ, tag)
	}

	w.line("}")
	w.line()
	return nil
}

func (w *writer) readStruct(def *compiler.Definition) error {
	w.linef(`func Read%v(b []byte) (str %v, err error) {`, def.Name, def.Name)
	w.line(`if len(b) == 0 {
		return
	}`)
	w.line(`err = str.Unmarshal(b)`)
	w.line(`return`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) structMarshal(def *compiler.Definition) error {
	w.linef(`func (s %v) Marshal() ([]byte, error) {`, def.Name)
	w.line(`return spec.Write(s)
	}`)
	w.line()

	w.linef(`func (s %v) MarshalTo(b []byte) ([]byte, error) {`, def.Name)
	w.line(`return spec.WriteTo(s, b)
	}`)
	w.line()
	return nil
}

func (w *writer) structUnmarshal(def *compiler.Definition) error {
	w.linef(`func (s *%v) Unmarshal(b []byte) error {`, def.Name)
	w.linef(`r, err := spec.ReadStruct(b)
	switch {
	case err != nil:
		return err
	case len(r) == 0:
		return nil
	}`)
	w.line()

	// unmarshal fields in reverse
	fields := def.Struct.Fields
	for i := len(fields) - 1; i >= 0; i-- {
		field := fields[i]
		if err := w.structUnmarshalField(field); err != nil {
			return err
		}
	}

	w.line(`return nil`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) structUnmarshalField(field *compiler.StructField) error {
	name := structFieldName(field)
	typ := field.Type
	kind := typ.Kind

	switch kind {
	case compiler.KindBool:
		w.linef(`s.%v, r, err = r.ReadBool()`, name)
		w.linef(`if err != nil {
			return err
		}`)

	case compiler.KindInt8:
		w.linef(`s.%v, r, err = r.ReadInt8()`, name)
		w.linef(`if err != nil {
			return err
		}`)
	case compiler.KindInt16:
		w.linef(`s.%v, r, err = r.ReadInt16()`, name)
		w.linef(`if err != nil {
			return err
		}`)
	case compiler.KindInt32:
		w.linef(`s.%v, r, err = r.ReadInt32()`, name)
		w.linef(`if err != nil {
			return err
		}`)
	case compiler.KindInt64:
		w.linef(`s.%v, r, err = r.ReadInt64()`, name)
		w.linef(`if err != nil {
			return err
		}`)

	case compiler.KindUint8:
		w.linef(`s.%v, r, err = r.ReadUint8()`, name)
		w.linef(`if err != nil {
			return err
		}`)
	case compiler.KindUint16:
		w.linef(`s.%v, r, err = r.ReadUint16()`, name)
		w.linef(`if err != nil {
			return err
		}`)
	case compiler.KindUint32:
		w.linef(`s.%v, r, err = r.ReadUint32()`, name)
		w.linef(`if err != nil {
			return err
		}`)
	case compiler.KindUint64:
		w.linef(`s.%v, r, err = r.ReadUint64()`, name)
		w.linef(`if err != nil {
			return err
		}`)

	case compiler.KindU128:
		w.linef(`s.%v, r, err = r.ReadU128()`, name)
		w.linef(`if err != nil {
			return err
		}`)
	case compiler.KindU256:
		w.linef(`s.%v, r, err = r.ReadU256()`, name)
		w.linef(`if err != nil {
			return err
		}`)

	case compiler.KindFloat32:
		w.linef(`s.%v, r, err = r.ReadFloat32()`, name)
		w.linef(`if err != nil {
			return err
		}`)
	case compiler.KindFloat64:
		w.linef(`s.%v, r, err = r.ReadFloat64()`, name)
		w.linef(`if err != nil {
			return err
		}`)

	case compiler.KindBytes:
		w.linef(`s.%v, r, err = r.ReadBytes()`, name)
		w.linef(`if err != nil {
			return err
		}`)
	case compiler.KindString:
		w.linef(`s.%v, r, err = r.ReadString()`, name)
		w.linef(`if err != nil {
			return err
		}`)

	case compiler.KindStruct:
		read := structReadFunc(typ)

		w.linef(`{`)
		w.linef(`data, r, err := r.Read()`)
		w.linef(`if err != nil {
			return err
		}`)

		w.linef(`s.%v, err = %v(data)`, name, read)
		w.linef(`if err != nil {
			return err
		}`)
		w.linef(`}`)
	}

	return nil
}

func (w *writer) structWrite(def *compiler.Definition) error {
	w.linef(`func (s %v) Write(w *spec.Writer) error {`, def.Name)
	w.linef(`if err := w.BeginStruct(); err != nil {
		return err
	}`)
	w.line()

	for _, field := range def.Struct.Fields {
		name := structFieldName(field)
		typ := field.Type
		val := fmt.Sprintf("s.%v", name)

		w.writerWrite(typ, val)
		w.linef(`w.StructField()`)
		w.line()
	}

	// end
	w.line(`return w.EndStruct()`)
	w.line(`}`)
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
