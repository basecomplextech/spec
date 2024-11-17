// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"math"
	"testing"

	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DecodeType

func TestDecodeType__should_return_type(t *testing.T) {
	b := []byte{}
	b = append(b, byte(core.TypeString))

	v, n, err := DecodeType(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, v, core.TypeString)
}

func TestDecodeType__should_return_undefined_when_empty(t *testing.T) {
	b := []byte{}

	v, n, err := DecodeType(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Zero(t, n)
	assert.Equal(t, v, core.TypeUndefined)
}

// DecodeBool

func TestDecodeBool__should_decode_bool_value(t *testing.T) {
	b := []byte{byte(core.TypeTrue)}
	v, n, err := DecodeBool(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, true, v)

	b = []byte{byte(core.TypeFalse)}
	v, n, err = DecodeBool(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, false, v)

	typ, size, err := DecodeTypeSize(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeFalse, typ)
	assert.Equal(t, size, len(b))
}

// DecodeByte

func TestDecodeByte__should_decode_byte(t *testing.T) {
	b := buffer.New()
	EncodeByte(b, 1)
	p := b.Bytes()

	v, n, err := DecodeByte(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, byte(1), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeByte, typ)
	assert.Equal(t, size, len(p))
}

// Int16

func TestDecodeInt16__should_decode_int16(t *testing.T) {
	b := buffer.New()
	EncodeInt16(b, math.MaxInt16)
	p := b.Bytes()

	v, n, err := DecodeInt16(p)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, b.Len())
	assert.Equal(t, int16(math.MaxInt16), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeInt16, typ)
	assert.Equal(t, size, len(p))
}

// Int32

func TestDecodeInt32__should_decode_int32(t *testing.T) {
	b := buffer.New()
	EncodeInt32(b, math.MaxInt32)
	p := b.Bytes()

	v, n, err := DecodeInt32(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, int32(math.MaxInt32), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeInt32, typ)
	assert.Equal(t, size, len(p))
}

// Int64

func TestDecodeInt64__should_decode_int64(t *testing.T) {
	b := buffer.New()
	EncodeInt64(b, math.MaxInt64)
	p := b.Bytes()

	v, n, err := DecodeInt64(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, int64(math.MaxInt64), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeInt64, typ)
	assert.Equal(t, size, len(p))
}

func TestDecodeInt64__should_decode_int64_from_int32(t *testing.T) {
	b := buffer.New()
	EncodeInt32(b, math.MaxInt32)
	p := b.Bytes()

	v, n, err := DecodeInt64(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, int64(math.MaxInt32), v)
}

// Uint16

func TestDecodeUint16__should_decode_int16(t *testing.T) {
	b := buffer.New()
	EncodeUint16(b, math.MaxUint16)
	p := b.Bytes()

	v, n, err := DecodeUint16(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, uint16(math.MaxUint16), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeUint16, typ)
	assert.Equal(t, size, len(p))
}

// Uint32

func TestDecodeUint32__should_decode_int32(t *testing.T) {
	b := buffer.New()
	EncodeUint32(b, math.MaxUint32)
	p := b.Bytes()

	v, n, err := DecodeUint32(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, uint32(math.MaxUint32), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeUint32, typ)
	assert.Equal(t, size, len(p))
}

// Uint64

func TestDecodeUint64__should_decode_int64(t *testing.T) {
	b := buffer.New()
	EncodeUint64(b, math.MaxUint64)
	p := b.Bytes()

	v, n, err := DecodeUint64(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, uint64(math.MaxUint64), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeUint64, typ)
	assert.Equal(t, size, len(p))
}

func TestDecodeUint64__should_decode_uint64_from_uint32(t *testing.T) {
	b := buffer.New()
	EncodeUint32(b, math.MaxUint32)
	p := b.Bytes()

	v, n, err := DecodeUint64(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, uint64(math.MaxUint32), v)
}

// Float32

func TestDecodeFloat32__should_decode_float32(t *testing.T) {
	b := buffer.New()
	EncodeFloat32(b, math.MaxFloat32)
	p := b.Bytes()

	v, n, err := DecodeFloat32(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, float32(math.MaxFloat32), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeFloat32, typ)
	assert.Equal(t, size, len(p))
}

func TestDecodeFloat32__should_decode_float32_from_float64(t *testing.T) {
	b := buffer.New()
	EncodeFloat64(b, math.MaxFloat32)
	p := b.Bytes()

	v, n, err := DecodeFloat32(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, float32(math.MaxFloat32), v)
}

// Float64

func TestDecodeFloat64__should_decode_float64(t *testing.T) {
	b := buffer.New()
	EncodeFloat64(b, math.MaxFloat64)
	p := b.Bytes()

	v, n, err := DecodeFloat64(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, float64(math.MaxFloat64), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeFloat64, typ)
	assert.Equal(t, size, len(p))
}

func TestDecodeFloat64__should_decode_float64_from_float32(t *testing.T) {
	b := buffer.New()
	EncodeFloat32(b, math.MaxFloat32)
	p := b.Bytes()

	v, n, err := DecodeFloat64(p)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, b.Len())
	assert.Equal(t, float64(math.MaxFloat32), v)
}

// Bin64/128/256

func TestDecodeBin64__should_decode_bin64(t *testing.T) {
	b := buffer.New()
	v := bin.Random64()
	EncodeBin64(b, v)
	p := b.Bytes()

	v1, n, err := DecodeBin64(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, v, v1)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeBin64, typ)
	assert.Equal(t, size, len(p))
}

func TestDecodeBin128__should_decode_bin128(t *testing.T) {
	b := buffer.New()
	v := bin.Random128()
	EncodeBin128(b, v)
	p := b.Bytes()

	v1, n, err := DecodeBin128(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, v, v1)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeBin128, typ)
	assert.Equal(t, size, len(p))
}

func TestDecodeBin256__should_decode_bin256(t *testing.T) {
	b := buffer.New()
	v := bin.Random256()
	EncodeBin256(b, v)
	p := b.Bytes()

	v1, n, err := DecodeBin256(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, v, v1)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeBin256, typ)
	assert.Equal(t, size, len(p))
}

// Bytes

func TestDecodeBytes__should_decode_bytes(t *testing.T) {
	v := []byte("hello, world")

	b := buffer.New()
	_, err := EncodeBytes(b, v)
	if err != nil {
		t.Fatal(err)
	}
	p := b.Bytes()

	v1, n, err := DecodeBytes(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, v, v1.Unwrap())

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeBytes, typ)
	assert.Equal(t, size, len(p))
}

// String

func TestDecodeString__should_decode_string(t *testing.T) {
	v := "hello, world"

	b := buffer.New()
	_, err := EncodeString(b, v)
	if err != nil {
		t.Fatal(err)
	}
	p := b.Bytes()

	v1, n, err := DecodeString(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, v, v1.Unwrap())

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeString, typ)
	assert.Equal(t, size, len(p))
}

// List

func TestDecodeListTable__should_decode_list(t *testing.T) {
	elements := TestElements()
	dataSize := 100
	b := testEncodeListTable(t, dataSize, elements)

	table, n, err := DecodeListTable(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(b), n)
	assert.Equal(t, uint32(dataSize), table.DataSize())
	assert.Equal(t, len(elements), table.Len())

	typ, size, err := DecodeTypeSize(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeList, typ)
	assert.Equal(t, size, len(b))
}

func TestDecodeListTable__should_decode_list_table(t *testing.T) {
	elements := TestElements()

	for i := 0; i <= len(elements); i++ {
		b := buffer.New()
		ee0 := elements[i:]

		size, err := encodeListTable(b, ee0, false)
		if err != nil {
			t.Fatal(err)
		}
		p := b.Bytes()

		table1, err := decodeListTable(p, uint32(size), false)
		if err != nil {
			t.Fatal(err)
		}

		ee1 := table1.elements(false)
		require.Equal(t, ee0, ee1)
	}
}

func TestDecodeListTable__should_return_error_when_invalid_type(t *testing.T) {
	elements := TestElements()
	dataSize := 100

	b := testEncodeListTable(t, dataSize, elements)
	b[len(b)-1] = byte(core.TypeMessage)

	_, _, err := DecodeListTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid type")
}

func TestDecodeListTable__should_return_error_when_invalid_table_size(t *testing.T) {
	b := []byte{}
	b = append(b, 0xff)
	b = append(b, byte(core.TypeList))

	_, _, err := DecodeListTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table size")
}

func TestDecodeListTable__should_return_error_when_invalid_data_size(t *testing.T) {
	big := false
	b := []byte{}
	b = append(b, 0xff)
	b = appendSize(b, big, 1000)
	b = append(b, byte(core.TypeList))

	_, _, err := DecodeListTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid data size")
}

func TestDecodeListTable__should_return_error_when_invalid_table(t *testing.T) {
	buf := buffer.New()
	_, err := encodeListTable(buf, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	big := false
	b := buf.Bytes()
	b = appendSize(b, big, 0)    // data size
	b = appendSize(b, big, 1000) // table size
	b = append(b, byte(core.TypeList))

	_, _, err = DecodeListTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table")
}

func TestDecodeListTable__should_return_error_when_invalid_data(t *testing.T) {
	buf := buffer.New()
	_, err := encodeListTable(buf, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	big := false
	b := buf.Bytes()
	b = appendSize(b, big, 1000) // data size
	b = appendSize(b, big, 0)    // table size
	b = append(b, byte(core.TypeList))

	_, _, err = DecodeListTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid data")
}

// Message

func TestDecodeMessageTable__should_decode_message_meta(t *testing.T) {
	fields := TestFields()
	dataSize := 100
	b := testEncodeMessageTable(t, dataSize, fields)

	meta, n, err := DecodeMessageTable(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(b), n)
	assert.Equal(t, uint32(dataSize), meta.DataSize())
	assert.Equal(t, len(fields), meta.Len())

	typ, size, err := DecodeTypeSize(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.TypeMessage, typ)
	assert.Equal(t, size, len(b))
}

func TestDecodeMessageTable__should_decode_message_table(t *testing.T) {
	fields := TestFields()

	for i := 0; i <= len(fields); i++ {
		buf := buffer.New()
		fields0 := fields[i:]

		size, err := encodeMessageTable(buf, fields0, false)
		if err != nil {
			t.Fatal(err)
		}

		table1, err := decodeMessageTable(buf.Bytes(), uint32(size), false)
		if err != nil {
			t.Fatal(err)
		}

		fields1 := table1.fields(false)
		require.Equal(t, fields0, fields1)
	}
}

func TestDecodeMessageTable__should_return_error_when_invalid_type(t *testing.T) {
	fields := TestFields()
	dataSize := 100

	b := testEncodeMessageTable(t, dataSize, fields)
	b[len(b)-1] = byte(core.TypeList)

	_, _, err := DecodeMessageTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid type")
}

func TestDecodeMessageTable__should_return_error_when_invalid_table_size(t *testing.T) {
	b := []byte{}
	b = append(b, 0xff)
	b = append(b, byte(core.TypeMessage))

	_, _, err := DecodeMessageTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table size")
}

func TestDecodeMessageTable__should_return_error_when_invalid_data_size(t *testing.T) {
	big := false
	b := []byte{}
	b = append(b, 0xff)
	b = appendSize(b, big, 1000)
	b = append(b, byte(core.TypeMessage))

	_, _, err := DecodeMessageTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid data size")
}

func TestDecodeMessageTable__should_return_error_when_invalid_table(t *testing.T) {
	buf := buffer.New()
	_, err := encodeMessageTable(buf, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	big := false
	b := buf.Bytes()
	b = appendSize(b, big, 0)    // data size
	b = appendSize(b, big, 1000) // table size
	b = append(b, byte(core.TypeMessage))

	_, _, err = DecodeMessageTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table")
}

func TestDecodeMessageTable__should_return_error_when_invalid_data(t *testing.T) {
	buf := buffer.New()

	_, err := encodeMessageTable(buf, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	big := false
	b := buf.Bytes()
	b = appendSize(b, big, 1000)
	b = appendSize(b, big, 0)
	b = append(b, byte(core.TypeMessage))

	_, _, err = DecodeMessageTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid data")
}
