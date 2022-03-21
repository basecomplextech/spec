package golang

import (
	"fmt"

	"github.com/complexl/spec/lang/compiler"
)

// dataGet returns a statement which accesses a data field, i.e. msgData.Int32(17).
func (w *writer) dataGet(typ *compiler.Type, data string, i string) string {
	kind := typ.Kind

	switch kind {
	default:
		panic(fmt.Sprintf("unsupported type kind %v", kind))

	case compiler.KindBool:
		return fmt.Sprintf(`%v.Bool(%v)`, data, i)

	case compiler.KindInt8:
		return fmt.Sprintf(`%v.Int8(%v)`, data, i)
	case compiler.KindInt16:
		return fmt.Sprintf(`%v.Int16(%v)`, data, i)
	case compiler.KindInt32:
		return fmt.Sprintf(`%v.Int32(%v)`, data, i)
	case compiler.KindInt64:
		return fmt.Sprintf(`%v.Int64(%v)`, data, i)

	case compiler.KindUint8:
		return fmt.Sprintf(`%v.Uint8(%v)`, data, i)
	case compiler.KindUint16:
		return fmt.Sprintf(`%v.Uint16(%v)`, data, i)
	case compiler.KindUint32:
		return fmt.Sprintf(`%v.Uint32(%v)`, data, i)
	case compiler.KindUint64:
		return fmt.Sprintf(`%v.Uint64(%v)`, data, i)

	case compiler.KindU128:
		return fmt.Sprintf(`%v.U128(%v)`, data, i)
	case compiler.KindU256:
		return fmt.Sprintf(`%v.U256(%v)`, data, i)

	case compiler.KindFloat32:
		return fmt.Sprintf(`%v.Float32(%v)`, data, i)
	case compiler.KindFloat64:
		return fmt.Sprintf(`%v.Float64(%v)`, data, i)

	case compiler.KindBytes:
		return fmt.Sprintf(`%v.Bytes(%v)`, data, i)
	case compiler.KindString:
		return fmt.Sprintf(`%v.String(%v)`, data, i)

	// list

	case compiler.KindList:
		panic("cannot access list as data")

	// resolved

	case compiler.KindEnum:
		readFunc := enumReadFunc(typ)
		w.linef(`v, _ := %v(%v)`, readFunc, data)
		return "v"

	case compiler.KindMessage:
		dataFunc := messageNewDataFunc(typ)
		w.linef(`v, _ := %v(%v)`, dataFunc, data)
		return "v"

	case compiler.KindStruct:
		readFunc := structReadFunc(typ)
		w.linef(`v, _ := %v(%v)`, readFunc, data)
		return "v"
	}
}

// readerRead returns a statement which calls a reader method, i.e. w.ReadInt32(17).
func (w *writer) readerRead(typ *compiler.Type, reader string, i string) string {
	kind := typ.Kind

	switch kind {
	default:
		panic(fmt.Sprintf("unsupported type kind %v", kind))

	case compiler.KindBool:
		return fmt.Sprintf(`%v.ReadBool(%v)`, reader, i)

	case compiler.KindInt8:
		return fmt.Sprintf(`%v.ReadByte(%v)`, reader, i)
	case compiler.KindInt16:
		return fmt.Sprintf(`%v.ReadInt16(%v)`, reader, i)
	case compiler.KindInt32:
		return fmt.Sprintf(`%v.ReadInt32(%v)`, reader, i)
	case compiler.KindInt64:
		return fmt.Sprintf(`%v.ReadInt64(%v)`, reader, i)

	case compiler.KindUint8:
		return fmt.Sprintf(`%v.ReadUint8(%v)`, reader, i)
	case compiler.KindUint16:
		return fmt.Sprintf(`%v.ReadUint16(%v)`, reader, i)
	case compiler.KindUint32:
		return fmt.Sprintf(`%v.ReadUint32(%v)`, reader, i)
	case compiler.KindUint64:
		return fmt.Sprintf(`%v.ReadUint64(%v)`, reader, i)

	case compiler.KindU128:
		return fmt.Sprintf(`%v.ReadU128(%v)`, reader, i)
	case compiler.KindU256:
		return fmt.Sprintf(`%v.ReadU256(%v)`, reader, i)

	case compiler.KindFloat32:
		return fmt.Sprintf(`%v.ReadFloat32(%v)`, reader, i)
	case compiler.KindFloat64:
		return fmt.Sprintf(`%v.ReadFloat64(%v)`, reader, i)

	case compiler.KindBytes:
		return fmt.Sprintf(`%v.ReadBytes(%v)`, reader, i)
	case compiler.KindString:
		return fmt.Sprintf(`%v.ReadString(%v)`, reader, i)

	// list

	case compiler.KindList:
		elem := typ.Element
		elemType := entryType(elem)

		// begin
		w.linef(`list, err := %v.DecodeList(%v)`, reader, i)
		w.linef(`if err != nil {
			return err
		}`)
		w.linef(`count := list.Len()`)
		w.linef(`elems := make([]%v, 0, count)`, elemType)
		w.line()

		// elements
		w.linef(`for i := 0; i < count; i++ {`)
		stmt := w.readerRead(elem, "list", "i")
		w.linef(`elem, err := %v`, stmt)
		w.linef(`if err != nil {
			return err
		}`)
		w.linef(`elems = append(elems, elem)`)
		w.line(`}`)
		w.line()

		return fmt.Sprintf("elems, nil")

	// resolved

	case compiler.KindEnum:
		readFunc := enumReadFunc(typ)

		w.linef(`data, err := %v.Read(%v)`, reader, i)
		w.linef(`if err != nil {
			return err
		}`)

		return fmt.Sprintf(`%v(data)`, readFunc)

	case compiler.KindMessage:
		readFunc := messageReadFunc(typ)

		w.linef(`data, err := %v.Read(%v)`, reader, i)
		w.linef(`if err != nil {
			return err
		}`)

		return fmt.Sprintf(`%v(data)`, readFunc)

	case compiler.KindStruct:
		readFunc := structReadFunc(typ)

		w.linef(`data, err := %v.Read(%v)`, reader, i)
		w.linef(`if err != nil {
			return err
		}`)

		return fmt.Sprintf(`%v(data)`, readFunc)
	}
}

// writerWrite writes a value to a write, i.e. w.Int32(m.Index).
func (w *writer) writerWrite(typ *compiler.Type, val string) error {
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

	case compiler.KindU128:
		w.linef(`w.U128(%v)`, val)
	case compiler.KindU256:
		w.linef(`w.U256(%v)`, val)

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
		w.linef(`if err := %v.Write(w); err != nil {`, val)
		w.linef(`return err`)
		w.line(`}`)

	case compiler.KindMessage:
		w.linef(`if err := %v.Write(w); err != nil {`, val)
		w.linef(`return err`)
		w.line(`}`)

	case compiler.KindStruct:
		w.linef(`if err := %v.Write(w); err != nil {`, val)
		w.linef(`return err`)
		w.linef(`}`)
	}
	return nil
}
