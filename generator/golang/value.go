package golang

import (
	"fmt"

	"github.com/baseone-run/spec/compiler"
)

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
