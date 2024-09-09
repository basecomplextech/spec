// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package model

import (
	"fmt"
	"strconv"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type Kind int

const (
	KindUndefined Kind = iota
	KindAny

	KindBool
	KindByte

	KindInt16
	KindInt32
	KindInt64

	KindUint16
	KindUint32
	KindUint64

	KindBin64
	KindBin128
	KindBin256

	KindFloat32
	KindFloat64

	KindBytes
	KindString
	KindAnyMessage

	// List

	KindList

	// Resolved

	KindEnum
	KindMessage
	KindStruct

	// Service

	KindService

	// Pending

	KindReference
)

func parseKind(pkind syntax.Kind) (Kind, error) {
	switch pkind {
	case syntax.KindAny:
		return KindAny, nil

	case syntax.KindBool:
		return KindBool, nil
	case syntax.KindByte:
		return KindByte, nil

	case syntax.KindInt16:
		return KindInt16, nil
	case syntax.KindInt32:
		return KindInt32, nil
	case syntax.KindInt64:
		return KindInt64, nil

	case syntax.KindUint16:
		return KindUint16, nil
	case syntax.KindUint32:
		return KindUint32, nil
	case syntax.KindUint64:
		return KindUint64, nil

	case syntax.KindBin64:
		return KindBin64, nil
	case syntax.KindBin128:
		return KindBin128, nil
	case syntax.KindBin256:
		return KindBin256, nil

	case syntax.KindFloat32:
		return KindFloat32, nil
	case syntax.KindFloat64:
		return KindFloat64, nil

	case syntax.KindBytes:
		return KindBytes, nil
	case syntax.KindString:
		return KindString, nil
	case syntax.KindAnyMessage:
		return KindAnyMessage, nil

	case syntax.KindList:
		return KindList, nil

	case syntax.KindReference:
		return KindReference, nil
	}

	return 0, fmt.Errorf("unknown type kind %v", pkind)
}

func (k Kind) String() string {
	switch k {
	case KindAny:
		return "any"

	case KindBool:
		return "bool"
	case KindByte:
		return "byte"

	case KindInt16:
		return "int16"
	case KindInt32:
		return "int32"
	case KindInt64:
		return "int64"

	case KindUint16:
		return "uint16"
	case KindUint32:
		return "uint32"
	case KindUint64:
		return "uint64"

	case KindBin64:
		return "bin64"
	case KindBin128:
		return "bin128"
	case KindBin256:
		return "bin256"

	case KindFloat32:
		return "float32"
	case KindFloat64:
		return "float64"

	case KindBytes:
		return "bytes"
	case KindString:
		return "string"
	case KindAnyMessage:
		return "message"

	case KindList:
		return "list"

	case KindEnum:
		return "enum"
	case KindMessage:
		return "message"
	case KindStruct:
		return "struct"

	case KindService:
		return "service"

	case KindReference:
		return "reference"
	}

	return strconv.Itoa(int(k))
}
