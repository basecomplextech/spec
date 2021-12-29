package generator

import (
	"fmt"

	"github.com/baseone-run/spec/compiler"
)

func (w *goWriter) writeValue(typ *compiler.Type, val string) error {
	kind := typ.Kind

	switch kind {
	default:
		panic(fmt.Sprintf("unsupported type kind %v", kind))

	case compiler.KindBool:
		w.linef(`w.Bool(%v)`, val)

	case compiler.KindInt8:
		w.linef(`w.Int8(%v)`, val)
	case compiler.KindInt16:
		w.linef(`w.Int16(%v)`, val)
	case compiler.KindInt32:
		w.linef(`w.Int32(%v)`, val)
	case compiler.KindInt64:
		w.linef(`w.Int64(%v)`, val)

	case compiler.KindUint8:
		w.linef(`w.Uint8(%v)`, val)
	case compiler.KindUint16:
		w.linef(`w.Uint16(%v)`, val)
	case compiler.KindUint32:
		w.linef(`w.Uint32(%v)`, val)
	case compiler.KindUint64:
		w.linef(`w.Uint64(%v)`, val)

	case compiler.KindFloat32:
		w.linef(`w.Float32(%v)`, val)
	case compiler.KindFloat64:
		w.linef(`w.Float64(%v)`, val)

	case compiler.KindBytes:
		w.linef(`w.Bytes(%v)`, val)
	case compiler.KindString:
		w.linef(`w.String(%v)`, val)

	// list

	case compiler.KindList:
		panic("cannot write list as value")

	// resolved

	case compiler.KindEnum:
		w.linef(`w.Int32(int32(%v))`, val)
	case compiler.KindMessage:
		w.linef(`if err := %v.Write(w); err != nil { return err }`, val)
	case compiler.KindStruct:
		w.linef(`if err := %v.Write(w); err != nil { return err }`, val)
	}
	return nil
}

func goTypeName(t *compiler.Type) string {
	switch t.Kind {
	case compiler.KindBool:
		return "bool"

	case compiler.KindInt8:
		return "int8"
	case compiler.KindInt16:
		return "int16"
	case compiler.KindInt32:
		return "int32"
	case compiler.KindInt64:
		return "int64"

	case compiler.KindUint8:
		return "uint8"
	case compiler.KindUint16:
		return "uint16"
	case compiler.KindUint32:
		return "uint32"
	case compiler.KindUint64:
		return "uint64"

	case compiler.KindFloat32:
		return "float32"
	case compiler.KindFloat64:
		return "float64"

	case compiler.KindBytes:
		return "[]byte"
	case compiler.KindString:
		return "string"

	// list

	case compiler.KindList:
		elem := goTypeName(t.Element)
		return "[]" + elem

	// resolved

	case compiler.KindEnum:
		if t.Import != nil {
			return fmt.Sprintf("%v.%v", t.ImportName, t.Name)
		}
		return t.Name

	case compiler.KindMessage:
		if t.Import != nil {
			return fmt.Sprintf("*%v.%v", t.ImportName, t.Name)
		}
		return "*" + t.Name

	case compiler.KindStruct:
		if t.Import != nil {
			return fmt.Sprintf("%v.%v", t.ImportName, t.Name)
		}
		return t.Name
	}

	return ""
}

func goReadFunc(t *compiler.Type) string {
	switch t.Kind {
	case compiler.KindMessage:
		if t.ImportName == "" {
			return fmt.Sprintf("Read%v", t.Name)
		} else {
			return fmt.Sprintf("%v.Read%v", t.ImportName, t.Name)
		}
	case compiler.KindStruct:
		if t.ImportName == "" {
			return fmt.Sprintf("Read%v", t.Name)
		} else {
			return fmt.Sprintf("%v.Read%v", t.ImportName, t.Name)
		}
	}

	panic(fmt.Sprintf("unsupported type kind %v", t.Kind))
}
