// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import "github.com/basecomplextech/spec/internal/decode"

var (
	DecodeType     = decode.DecodeType
	DecodeTypeSize = decode.DecodeTypeSize

	DecodeBool = decode.DecodeBool
	DecodeByte = decode.DecodeByte

	DecodeBin64  = decode.DecodeBin64
	DecodeBin128 = decode.DecodeBin128
	DecodeBin256 = decode.DecodeBin256

	DecodeBytes = decode.DecodeBytes

	DecodeFloat32 = decode.DecodeFloat32
	DecodeFloat64 = decode.DecodeFloat64

	DecodeInt16 = decode.DecodeInt16
	DecodeInt32 = decode.DecodeInt32
	DecodeInt64 = decode.DecodeInt64

	DecodeListTable    = decode.DecodeListTable
	DecodeMessageTable = decode.DecodeMessageTable

	DecodeString      = decode.DecodeString
	DecodeStringClone = decode.DecodeStringClone

	DecodeStruct = decode.DecodeStruct

	DecodeUint16 = decode.DecodeUint16
	DecodeUint32 = decode.DecodeUint32
	DecodeUint64 = decode.DecodeUint64
)
