// Copyright 2021 Ivan Korobkov. All rights reserved.

package generator

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/model"
)

// typeName returns a type name.
func typeName(typ *model.Type) string {
	kind := typ.Kind

	switch kind {
	case model.KindAny:
		return "spec.Value"

	case model.KindBool:
		return "bool"
	case model.KindByte:
		return "byte"

	case model.KindInt16:
		return "int16"
	case model.KindInt32:
		return "int32"
	case model.KindInt64:
		return "int64"

	case model.KindUint16:
		return "uint16"
	case model.KindUint32:
		return "uint32"
	case model.KindUint64:
		return "uint64"

	case model.KindFloat32:
		return "float32"
	case model.KindFloat64:
		return "float64"

	case model.KindBin64:
		return "bin.Bin64"
	case model.KindBin128:
		return "bin.Bin128"
	case model.KindBin256:
		return "bin.Bin256"

	case model.KindBytes:
		return "[]byte"
	case model.KindString:
		return "string"
	case model.KindAnyMessage:
		return "spec.Message"

	case model.KindList:
		elem := typeName(typ.Element)
		return fmt.Sprintf("spec.TypedList[%v]", elem)

	case model.KindEnum,
		model.KindMessage,
		model.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.%v", typ.ImportName, typ.Name)
		}
		return typ.Name

	case model.KindService:
		if typ.Import != nil {
			return fmt.Sprintf("%v.%v", typ.ImportName, typ.Name)
		}
		return typ.Name
	}

	panic(fmt.Sprintf("unsupported type kind %v", typ.Kind))
}

func typeRefName(typ *model.Type) string {
	kind := typ.Kind

	switch kind {
	case model.KindBytes:
		return "spec.Bytes"
	case model.KindString:
		return "spec.String"

	case model.KindList:
		elem := typeRefName(typ.Element)
		return fmt.Sprintf("spec.TypedList[%v]", elem)
	}

	return typeName(typ)
}

func inTypeName(typ *model.Type) string {
	kind := typ.Kind
	switch kind {
	case model.KindBytes:
		return "[]byte"
	case model.KindString:
		return "string"
	}
	return typeName(typ)
}

func typeNewFunc(typ *model.Type) string {
	kind := typ.Kind

	switch kind {
	case model.KindList:
		elem := typeName(typ.Element)
		return "spec.List[]" + elem

	case model.KindEnum,
		model.KindMessage,
		model.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.New%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("New%v", typ.Name)
	}
	return ""
}

func typeMakeMessageFunc(typ *model.Type) string {
	kind := typ.Kind

	switch kind {
	case model.KindMessage:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Make%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Make%v", typ.Name)
	}

	panic(fmt.Sprintf("unsupported type kind %v", typ.Kind))
}

func typeParseFunc(typ *model.Type) string {
	kind := typ.Kind

	switch kind {
	case model.KindList:
		elem := typeName(typ.Element)
		return "spec.List[]" + elem

	case model.KindEnum,
		model.KindMessage,
		model.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Parse%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Parse%v", typ.Name)

	default:
		return typeDecodeFunc(typ)
	}
	return ""
}

func typeDecodeFunc(typ *model.Type) string {
	kind := typ.Kind

	switch kind {
	case model.KindAny:
		return "spec.ParseValue"

	case model.KindBool:
		return "encoding.DecodeBool"
	case model.KindByte:
		return "encoding.DecodeByte"

	case model.KindInt16:
		return "encoding.DecodeInt16"
	case model.KindInt32:
		return "encoding.DecodeInt32"
	case model.KindInt64:
		return "encoding.DecodeInt64"

	case model.KindUint16:
		return "encoding.DecodeUint16"
	case model.KindUint32:
		return "encoding.DecodeUint32"
	case model.KindUint64:
		return "encoding.DecodeUint64"

	case model.KindBin64:
		return "encoding.DecodeBin64"
	case model.KindBin128:
		return "encoding.DecodeBin128"
	case model.KindBin256:
		return "encoding.DecodeBin256"

	case model.KindFloat32:
		return "encoding.DecodeFloat32"
	case model.KindFloat64:
		return "encoding.DecodeFloat64"

	case model.KindBytes:
		return "encoding.DecodeBytes"
	case model.KindString:
		return "encoding.DecodeString"
	case model.KindAnyMessage:
		return "spec.ParseMessage"

	case model.KindList:
		elem := typeName(typ.Element)
		return fmt.Sprintf("spec.ParseTypedList[%v]", elem)

	case model.KindEnum,
		model.KindMessage,
		model.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Parse%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Parse%v", typ.Name)
	}

	return ""
}

func typeDecodeRefFunc(typ *model.Type) string {
	kind := typ.Kind

	switch kind {
	case model.KindBytes:
		return "encoding.DecodeBytes"
	case model.KindString:
		return "encoding.DecodeString"

	case model.KindList:
		elem := typeRefName(typ.Element)
		return fmt.Sprintf("spec.ParseTypedList[%v]", elem)
	}

	return typeDecodeFunc(typ)
}

func typeWriteFunc(typ *model.Type) string {
	kind := typ.Kind

	switch kind {
	case model.KindAny:
		return "spec.WriteValue"

	case model.KindBool:
		return "encoding.EncodeBool"
	case model.KindByte:
		return "encoding.EncodeByte"

	case model.KindInt16:
		return "encoding.EncodeInt16"
	case model.KindInt32:
		return "encoding.EncodeInt32"
	case model.KindInt64:
		return "encoding.EncodeInt64"

	case model.KindUint16:
		return "encoding.EncodeUint16"
	case model.KindUint32:
		return "encoding.EncodeUint32"
	case model.KindUint64:
		return "encoding.EncodeUint64"

	case model.KindBin64:
		return "encoding.EncodeBin64"
	case model.KindBin128:
		return "encoding.EncodeBin128"
	case model.KindBin256:
		return "encoding.EncodeBin256"

	case model.KindFloat32:
		return "encoding.EncodeFloat32"
	case model.KindFloat64:
		return "encoding.EncodeFloat64"

	case model.KindBytes:
		return "encoding.EncodeBytes"
	case model.KindString:
		return "encoding.EncodeString"
	case model.KindAnyMessage:
		return "spec.WriteMessage"

	case model.KindEnum:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Write%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Write%v", typ.Name)

	case model.KindList:
		elem := typ.Element
		if elem.Kind == model.KindMessage {
			return fmt.Sprintf("spec.NewMessageListWriter")
		}
		return fmt.Sprintf("spec.NewValueListWriter")

	case model.KindMessage:
		if typ.Import != nil {
			return fmt.Sprintf("%v.New%vWriterTo", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("New%vWriterTo", typ.Name)

	case model.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Write%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Write%v", typ.Name)
	}

	return ""
}

func typeWriter(typ *model.Type) string {
	kind := typ.Kind

	switch kind {
	case model.KindList:
		elem := typ.Element
		if elem.Kind == model.KindMessage {
			encoder := typeWriter(elem)
			return fmt.Sprintf("spec.MessageListWriter[%v]", encoder)
		}

		elemName := inTypeName(elem)
		return fmt.Sprintf("spec.ValueListWriter[%v]", elemName)

	case model.KindMessage:
		if typ.Import != nil {
			return fmt.Sprintf("%v.%vWriter", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("%vWriter", typ.Name)
	}

	return ""
}
