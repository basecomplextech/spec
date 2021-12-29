package golang

import (
	"fmt"

	"github.com/baseone-run/spec/compiler"
)

func typeName(t *compiler.Type) string {
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
		elem := typeName(t.Element)
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

	panic(fmt.Sprintf("unsupported type kind %v", t.Kind))
}

func typeReadFunc(t *compiler.Type) string {
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
