// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package spec

import (
	"github.com/basecomplextech/spec/internal/encode"
	"github.com/basecomplextech/spec/internal/format"
)

type (
	ListTable    = format.ListTable
	MessageTable = format.MessageTable
)

var (
	EncodeBool = encode.EncodeBool
	EncodeByte = encode.EncodeByte

	EncodeBin64       = encode.EncodeBin64
	EncodeBin128      = encode.EncodeBin128
	EncodeBin128Bytes = encode.EncodeBin128Bytes
	EncodeBin256      = encode.EncodeBin256

	EncodeBytes = encode.EncodeBytes

	EncodeFloat32 = encode.EncodeFloat32
	EncodeFloat64 = encode.EncodeFloat64

	EncodeInt16 = encode.EncodeInt16
	EncodeInt32 = encode.EncodeInt32
	EncodeInt64 = encode.EncodeInt64

	EncodeListTable    = encode.EncodeListTable
	EncodeMessageTable = encode.EncodeMessageTable

	EncodeString = encode.EncodeString
	EncodeStruct = encode.EncodeStruct

	EncodeUint16 = encode.EncodeUint16
	EncodeUint32 = encode.EncodeUint32
	EncodeUint64 = encode.EncodeUint64
)
