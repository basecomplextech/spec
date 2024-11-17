// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

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
			return fmt.Sprintf("%v.Open%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Open%v", typ.Name)
	}
	return ""
}

func typeMakeMessageFunc(typ *model.Type) string {
	kind := typ.Kind

	switch kind {
	case model.KindMessage:
		if typ.Import != nil {
			return fmt.Sprintf("%v.New%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("New%v", typ.Name)
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
		model.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Decode%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Decode%v", typ.Name)

	case model.KindMessage:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Parse%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Parse%v", typ.Name)

	default:
		return typeDecodeFunc(typ)
	}
}

func typeDecodeFunc(typ *model.Type) string {
	kind := typ.Kind

	switch kind {
	case model.KindAny:
		return "spec.ParseValue"

	case model.KindBool:
		return "spec.DecodeBool"
	case model.KindByte:
		return "spec.DecodeByte"

	case model.KindInt16:
		return "spec.DecodeInt16"
	case model.KindInt32:
		return "spec.DecodeInt32"
	case model.KindInt64:
		return "spec.DecodeInt64"

	case model.KindUint16:
		return "spec.DecodeUint16"
	case model.KindUint32:
		return "spec.DecodeUint32"
	case model.KindUint64:
		return "spec.DecodeUint64"

	case model.KindBin64:
		return "spec.DecodeBin64"
	case model.KindBin128:
		return "spec.DecodeBin128"
	case model.KindBin256:
		return "spec.DecodeBin256"

	case model.KindFloat32:
		return "spec.DecodeFloat32"
	case model.KindFloat64:
		return "spec.DecodeFloat64"

	case model.KindBytes:
		return "spec.DecodeBytes"
	case model.KindString:
		return "spec.DecodeString"
	case model.KindAnyMessage:
		return "spec.ParseMessage"

	case model.KindList:
		elem := typeName(typ.Element)
		return fmt.Sprintf("spec.ParseTypedList[%v]", elem)

	case model.KindEnum,
		model.KindStruct:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Decode%v", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Decode%v", typ.Name)

	case model.KindMessage:
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
		return "spec.DecodeBytes"
	case model.KindString:
		return "spec.DecodeString"

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
		return "spec.EncodeBool"
	case model.KindByte:
		return "spec.EncodeByte"

	case model.KindInt16:
		return "spec.EncodeInt16"
	case model.KindInt32:
		return "spec.EncodeInt32"
	case model.KindInt64:
		return "spec.EncodeInt64"

	case model.KindUint16:
		return "spec.EncodeUint16"
	case model.KindUint32:
		return "spec.EncodeUint32"
	case model.KindUint64:
		return "spec.EncodeUint64"

	case model.KindBin64:
		return "spec.EncodeBin64"
	case model.KindBin128:
		return "spec.EncodeBin128"
	case model.KindBin256:
		return "spec.EncodeBin256"

	case model.KindFloat32:
		return "spec.EncodeFloat32"
	case model.KindFloat64:
		return "spec.EncodeFloat64"

	case model.KindBytes:
		return "spec.EncodeBytes"
	case model.KindString:
		return "spec.EncodeString"
	case model.KindAnyMessage:
		return "spec.WriteMessage"

	case model.KindEnum:
		if typ.Import != nil {
			return fmt.Sprintf("%v.Encode%vTo", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Encode%vTo", typ.Name)

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
			return fmt.Sprintf("%v.Encode%vTo", typ.ImportName, typ.Name)
		}
		return fmt.Sprintf("Encode%vTo", typ.Name)
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
