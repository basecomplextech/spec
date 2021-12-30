package golang

import (
	"fmt"

	"github.com/baseone-run/spec/compiler"
)

// dataType returns a data type name.
func dataType(typ *compiler.Type) string {
	return _typeName(typ, true)
}

// objectType returns an object type name.
func objectType(typ *compiler.Type) string {
	return _typeName(typ, false)
}

func _typeName(typ *compiler.Type, data bool) string {
	switch typ.Kind {
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
		elem := objectType(typ.Element)
		return "[]" + elem

	// resolved

	case compiler.KindEnum:
		if typ.Import != nil {
			return fmt.Sprintf("%v.%v", typ.ImportName, typ.Name)
		}
		return typ.Name

	case compiler.KindMessage:
		if data {
			if typ.Import != nil {
				return fmt.Sprintf("%v.%vData", typ.ImportName, typ.Name)
			} else {
				return fmt.Sprintf("%vData", typ.Name)
			}
		} else {
			if typ.Import != nil {
				return fmt.Sprintf("*%v.%v", typ.ImportName, typ.Name)
			} else {
				return fmt.Sprintf("*%v", typ.Name)
			}
		}

	case compiler.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.%v", typ.ImportName, typ.Name)
		}
		return typ.Name
	}

	panic(fmt.Sprintf("unsupported type kind %v", typ.Kind))
}
