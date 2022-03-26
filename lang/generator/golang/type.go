package golang

import (
	"fmt"

	"github.com/baseblck/spec/lang/compiler"
)

// typeName returns a type name.
func typeName(typ *compiler.Type) string {
	kind := typ.Kind

	switch kind {
	case compiler.KindBool:
		return "bool"
	case compiler.KindByte:
		return "byte"

	case compiler.KindInt32:
		return "int32"
	case compiler.KindInt64:
		return "int64"
	case compiler.KindUint32:
		return "uint32"
	case compiler.KindUint64:
		return "uint64"

	case compiler.KindU128:
		return "u128.U128"
	case compiler.KindU256:
		return "u256.U256"

	case compiler.KindFloat32:
		return "float32"
	case compiler.KindFloat64:
		return "float64"

	case compiler.KindBytes:
		return "[]byte"
	case compiler.KindString:
		return "string"

	case compiler.KindList:
		elem := typeName(typ.Element)
		return fmt.Sprintf("spec.List[%v]", elem)

	case compiler.KindEnum,
		compiler.KindMessage,
		compiler.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.%v", typ.ImportName, typ.Name)
		}
		return typ.Name
	}

	panic(fmt.Sprintf("unsupported type kind %v", typ.Kind))
}

func typeGetFunc(typ *compiler.Type) string {
	kind := typ.Kind

	switch kind {
	case compiler.KindList:
		elem := typeName(typ.Element)
		return "spec.List[]" + elem

	case compiler.KindEnum,
		compiler.KindMessage,
		compiler.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Get%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Get%v", typ.Name)
	}
	return ""
}

func typeDecodeFunc(typ *compiler.Type) string {
	kind := typ.Kind

	switch kind {
	case compiler.KindBool:
		return "spec.DecodeBool"
	case compiler.KindByte:
		return "spec.DecodeByte"

	case compiler.KindInt32:
		return "spec.DecodeInt32"
	case compiler.KindInt64:
		return "spec.DecodeInt64"
	case compiler.KindUint32:
		return "spec.DecodeUint32"
	case compiler.KindUint64:
		return "spec.DecodeUint64"

	case compiler.KindU128:
		return "spec.DecodeU128"
	case compiler.KindU256:
		return "spec.DecodeU256"

	case compiler.KindFloat32:
		return "spec.DecodeFloat32"
	case compiler.KindFloat64:
		return "spec.DecodeFloat64"

	case compiler.KindBytes:
		return "spec.DecodeBytes"
	case compiler.KindString:
		return "spec.DecodeString"

	case compiler.KindList:
		elem := typeName(typ.Element)
		return fmt.Sprintf("spec.DecodeList[%v]", elem)

	case compiler.KindEnum,
		compiler.KindMessage,
		compiler.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Decode%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Decode%v", typ.Name)
	}

	return ""
}

func typeEncodeFunc(typ *compiler.Type) string {
	kind := typ.Kind

	switch kind {
	case compiler.KindBool:
		return "spec.EncodeBool"
	case compiler.KindByte:
		return "spec.EncodeByte"

	case compiler.KindInt32:
		return "spec.EncodeInt32"
	case compiler.KindInt64:
		return "spec.EncodeInt64"
	case compiler.KindUint32:
		return "spec.EncodeUint32"
	case compiler.KindUint64:
		return "spec.EncodeUint64"

	case compiler.KindU128:
		return "spec.EncodeU128"
	case compiler.KindU256:
		return "spec.EncodeU256"

	case compiler.KindFloat32:
		return "spec.EncodeFloat32"
	case compiler.KindFloat64:
		return "spec.EncodeFloat64"

	case compiler.KindBytes:
		return "spec.EncodeBytes"
	case compiler.KindString:
		return "spec.EncodeString"

	case compiler.KindEnum,
		compiler.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Encode%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Encode%v", typ.Name)

	case compiler.KindList:
		elem := typ.Element
		if elem.Kind == compiler.KindMessage {
			return fmt.Sprintf("spec.BuildNestedList")
		}
		return fmt.Sprintf("spec.BuildList")

	case compiler.KindMessage:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Build%vEncoder", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Build%vEncoder", typ.Name)
	}

	return ""
}

func typeBuilder(typ *compiler.Type) string {
	kind := typ.Kind

	switch kind {
	case compiler.KindList:
		elem := typ.Element
		if elem.Kind == compiler.KindMessage {
			encoder := typeBuilder(elem)
			return fmt.Sprintf("spec.NestedListBuilder[%v]", encoder)
		}

		elemName := typeName(elem)
		return fmt.Sprintf("spec.ListBuilder[%v]", elemName)

	case compiler.KindMessage:
		if typ.Import != nil {
			return fmt.Sprintf("%v.%vBuilder", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("%vBuilder", typ.Name)
	}

	return ""
}
