package golang

import (
	"fmt"

	"github.com/baseone-run/spec/compiler"
)

func (w *writer) readValue(typ *compiler.Type, src string, tag string) string {
	kind := typ.Kind

	switch kind {
	default:
		panic(fmt.Sprintf("unsupported type kind %v", kind))

	case compiler.KindBool:
		return fmt.Sprintf(`%v.Bool(%v)`, src, tag)

	case compiler.KindInt8:
		return fmt.Sprintf(`%v.Int8(%v)`, src, tag)
	case compiler.KindInt16:
		return fmt.Sprintf(`%v.Int16(%v)`, src, tag)
	case compiler.KindInt32:
		return fmt.Sprintf(`%v.Int32(%v)`, src, tag)
	case compiler.KindInt64:
		return fmt.Sprintf(`%v.Int64(%v)`, src, tag)

	case compiler.KindUint8:
		return fmt.Sprintf(`%v.Uint8(%v)`, src, tag)
	case compiler.KindUint16:
		return fmt.Sprintf(`%v.Uint16(%v)`, src, tag)
	case compiler.KindUint32:
		return fmt.Sprintf(`%v.Uint32(%v)`, src, tag)
	case compiler.KindUint64:
		return fmt.Sprintf(`%v.Uint64(%v)`, src, tag)

	case compiler.KindFloat32:
		return fmt.Sprintf(`%v.Float32(%v)`, src, tag)
	case compiler.KindFloat64:
		return fmt.Sprintf(`%v.Float64(%v)`, src, tag)

	case compiler.KindBytes:
		return fmt.Sprintf(`%v.Bytes(%v)`, src, tag)
	case compiler.KindString:
		return fmt.Sprintf(`%v.String(%v)`, src, tag)

	// list

	case compiler.KindList:
		panic("cannot read list as value")

	// resolved

	case compiler.KindEnum:
		typeName := typeName(typ)
		return fmt.Sprintf(`%v(%v.Int32(%v))`, typeName, src, tag)

	case compiler.KindMessage:
		read := typeReadFunc(typ)
		return fmt.Sprintf(`%v(%v.Element(%v))`, read, src, tag)

	case compiler.KindStruct:
		read := typeReadFunc(typ)
		return fmt.Sprintf(`%v(%v.Element(%v))`, read, src, tag)
	}
}

func (w *writer) writeValue(typ *compiler.Type, val string) error {
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
