package golang

import (
	"fmt"

	"github.com/baseone-run/spec/compiler"
)

func (w *writer) listElementData(typ *compiler.Type) {
	w._listElement(typ, true)
}

func (w *writer) _listElement(typ *compiler.Type, data bool) {
	kind := typ.Kind

	switch kind {
	default:
		panic(fmt.Sprintf("unsupported type kind %v", kind))

	case compiler.KindBool:
		w.line(`elem := list.Bool(i)`)

	case compiler.KindInt8:
		w.line(`elem := list.Int8(i)`)
	case compiler.KindInt16:
		w.line(`elem := list.Int16(i)`)
	case compiler.KindInt32:
		w.line(`elem := list.Int32(i)`)
	case compiler.KindInt64:
		w.line(`elem := list.Int64(i)`)

	case compiler.KindUint8:
		w.line(`elem := list.Uint8(i)`)
	case compiler.KindUint16:
		w.line(`elem := list.Uint16(i)`)
	case compiler.KindUint32:
		w.line(`elem := list.Uint32(i)`)
	case compiler.KindUint64:
		w.line(`elem := list.Uint64(i)`)

	case compiler.KindFloat32:
		w.line(`elem := list.Float32(i)`)
	case compiler.KindFloat64:
		w.line(`elem := list.Float64(i)`)

	case compiler.KindBytes:
		w.line(`elem := list.Bytes(i)`)
	case compiler.KindString:
		w.line(`elem := list.String(i)`)

	// list

	case compiler.KindList:
		panic("cannot read list as list element")

	// resolved

	case compiler.KindEnum:
		objectType := objectType(typ)
		w.linef(`%v(list.Int32(i))`, objectType)

	case compiler.KindMessage:
		access := ""
		if data {
			access = messageNewDataFunc(typ)
		} else {
			access = messageReadFunc(typ)
		}

		w.linef(`data := list.Element(i)`)
		w.linef(`elem, err := %v(data)
		if err != nil {
			return err
		}`, access)

	case compiler.KindStruct:
		readFunc := structReadFunc(typ)
		w.linef(`data := list.Element(i)`)
		w.linef(`elem, err := %v(data)
		if err != nil {
			return err
		}`, readFunc)
	}
}
