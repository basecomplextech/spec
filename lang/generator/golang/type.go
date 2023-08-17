package golang

import (
	"fmt"

	"github.com/basecomplextech/spec/lang/compiler"
)

// typeName returns a type name.
func typeName(typ *compiler.Type) string {
	kind := typ.Kind

	switch kind {
	case compiler.KindAny:
		return "spec.Value"

	case compiler.KindBool:
		return "bool"
	case compiler.KindByte:
		return "byte"

	case compiler.KindInt16:
		return "int16"
	case compiler.KindInt32:
		return "int32"
	case compiler.KindInt64:
		return "int64"

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

	case compiler.KindBin64:
		return "bin.Bin64"
	case compiler.KindBin128:
		return "bin.Bin128"
	case compiler.KindBin256:
		return "bin.Bin256"

	case compiler.KindBytes:
		return "[]byte"
	case compiler.KindString:
		return "string"
	case compiler.KindAnyMessage:
		return "spec.Message"

	case compiler.KindList:
		elem := typeName(typ.Element)
		return fmt.Sprintf("spec.TypedList[%v]", elem)

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

func typeRefName(typ *compiler.Type) string {
	kind := typ.Kind

	switch kind {
	case compiler.KindBytes:
		return "spec.Bytes"
	case compiler.KindString:
		return "spec.String"

	case compiler.KindList:
		elem := typeRefName(typ.Element)
		return fmt.Sprintf("spec.TypedList[%v]", elem)
	}

	return typeName(typ)
}

func inTypeName(typ *compiler.Type) string {
	kind := typ.Kind
	switch kind {
	case compiler.KindBytes:
		return "[]byte"
	case compiler.KindString:
		return "string"
	}
	return typeName(typ)
}

func typeNewFunc(typ *compiler.Type) string {
	kind := typ.Kind

	switch kind {
	case compiler.KindList:
		elem := typeName(typ.Element)
		return "spec.List[]" + elem

	case compiler.KindEnum,
		compiler.KindMessage,
		compiler.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.New%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("New%v", typ.Name)
	}
	return ""
}

func typeDecodeFunc(typ *compiler.Type) string {
	kind := typ.Kind

	switch kind {
	case compiler.KindAny:
		return "spec.ParseValue"

	case compiler.KindBool:
		return "encoding.DecodeBool"
	case compiler.KindByte:
		return "encoding.DecodeByte"

	case compiler.KindInt16:
		return "encoding.DecodeInt16"
	case compiler.KindInt32:
		return "encoding.DecodeInt32"
	case compiler.KindInt64:
		return "encoding.DecodeInt64"

	case compiler.KindUint16:
		return "encoding.DecodeUint16"
	case compiler.KindUint32:
		return "encoding.DecodeUint32"
	case compiler.KindUint64:
		return "encoding.DecodeUint64"

	case compiler.KindBin64:
		return "encoding.DecodeBin64"
	case compiler.KindBin128:
		return "encoding.DecodeBin128"
	case compiler.KindBin256:
		return "encoding.DecodeBin256"

	case compiler.KindFloat32:
		return "encoding.DecodeFloat32"
	case compiler.KindFloat64:
		return "encoding.DecodeFloat64"

	case compiler.KindBytes:
		return "encoding.DecodeBytes"
	case compiler.KindString:
		return "encoding.DecodeString"
	case compiler.KindAnyMessage:
		return "spec.ParseMessage"

	case compiler.KindList:
		elem := typeName(typ.Element)
		return fmt.Sprintf("spec.ParseTypedList[%v]", elem)

	case compiler.KindEnum,
		compiler.KindMessage,
		compiler.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Parse%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Parse%v", typ.Name)
	}

	return ""
}

func typeDecodeRefFunc(typ *compiler.Type) string {
	kind := typ.Kind

	switch kind {
	case compiler.KindBytes:
		return "encoding.DecodeBytes"
	case compiler.KindString:
		return "encoding.DecodeString"

	case compiler.KindList:
		elem := typeRefName(typ.Element)
		return fmt.Sprintf("spec.ParseTypedList[%v]", elem)
	}

	return typeDecodeFunc(typ)
}

func typeWriteFunc(typ *compiler.Type) string {
	kind := typ.Kind

	switch kind {
	case compiler.KindAny:
		return "spec.WriteValue"

	case compiler.KindBool:
		return "encoding.EncodeBool"
	case compiler.KindByte:
		return "encoding.EncodeByte"

	case compiler.KindInt16:
		return "encoding.EncodeInt16"
	case compiler.KindInt32:
		return "encoding.EncodeInt32"
	case compiler.KindInt64:
		return "encoding.EncodeInt64"

	case compiler.KindUint16:
		return "encoding.EncodeUint16"
	case compiler.KindUint32:
		return "encoding.EncodeUint32"
	case compiler.KindUint64:
		return "encoding.EncodeUint64"

	case compiler.KindBin64:
		return "encoding.EncodeBin64"
	case compiler.KindBin128:
		return "encoding.EncodeBin128"
	case compiler.KindBin256:
		return "encoding.EncodeBin256"

	case compiler.KindFloat32:
		return "encoding.EncodeFloat32"
	case compiler.KindFloat64:
		return "encoding.EncodeFloat64"

	case compiler.KindBytes:
		return "encoding.EncodeBytes"
	case compiler.KindString:
		return "encoding.EncodeString"
	case compiler.KindAnyMessage:
		return "spec.WriteMessage"

	case compiler.KindEnum:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Write%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Write%v", typ.Name)

	case compiler.KindList:
		elem := typ.Element
		if elem.Kind == compiler.KindMessage {
			return fmt.Sprintf("spec.NewMessageListWriter")
		}
		return fmt.Sprintf("spec.NewValueListWriter")

	case compiler.KindMessage:
		if typ.Import != nil {
			return fmt.Sprintf("%v.New%vWriterTo", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("New%vWriterTo", typ.Name)

	case compiler.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Write%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Write%v", typ.Name)
	}

	return ""
}

func typeWriter(typ *compiler.Type) string {
	kind := typ.Kind

	switch kind {
	case compiler.KindList:
		elem := typ.Element
		if elem.Kind == compiler.KindMessage {
			encoder := typeWriter(elem)
			return fmt.Sprintf("spec.MessageListWriter[%v]", encoder)
		}

		elemName := inTypeName(elem)
		return fmt.Sprintf("spec.ValueListWriter[%v]", elemName)

	case compiler.KindMessage:
		if typ.Import != nil {
			return fmt.Sprintf("%v.%vWriter", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("%vWriter", typ.Name)
	}

	return ""
}
