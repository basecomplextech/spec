package spec

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DecodeType

func TestDecodeType__should_return_type(t *testing.T) {
	b := []byte{}
	b = append(b, byte(TypeString))

	v, n, err := DecodeType(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, v, TypeString)
}

func TestDecodeType__should_return_nil_when_empty(t *testing.T) {
	b := []byte{}

	v, n, err := DecodeType(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, v, TypeNil)
}

// DecodeBool

func TestDecodeBool__should_decode_bool_value(t *testing.T) {
	b := []byte{byte(TypeTrue)}
	v, n, err := DecodeBool(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, true, v)

	b = []byte{byte(TypeFalse)}
	v, n, err = DecodeBool(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, false, v)
}

// DecodeByte

func TestDecodeByte__should_decode_byte(t *testing.T) {
	b := []byte{}
	b = EncodeByte(b, 1)

	v, n, err := DecodeByte(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, byte(1), v)
}

func TestDecodeByte__should_decode_byte_from_int32(t *testing.T) {
	b := []byte{}
	b = EncodeInt32(b, 1)

	v, n, err := DecodeByte(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, byte(1), v)
}

func TestDecodeByte__should_decode_byte_from_int64(t *testing.T) {
	b := []byte{}
	b = EncodeInt64(b, 1)

	v, n, err := DecodeByte(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, byte(1), v)
}

// Int32

func TestDecodeInt32__should_decode_int32(t *testing.T) {
	b := []byte{}
	b = EncodeInt32(b, 1)

	v, n, err := DecodeInt32(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, int32(1), v)
}

// Int64

func TestDecodeInt64__should_decode_int64(t *testing.T) {
	b := []byte{}
	b = EncodeInt64(b, 1)

	v, n, err := DecodeInt64(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, int64(1), v)
}

func TestDecodeInt64__should_decode_int64_from_byte(t *testing.T) {
	b := []byte{}
	b = EncodeByte(b, 1)

	v, n, err := DecodeInt64(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, int64(1), v)
}

func TestDecodeInt64__should_decode_int64_from_int32(t *testing.T) {
	b := []byte{}
	b = EncodeInt32(b, 1)

	v, n, err := DecodeInt64(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, int64(1), v)
}

// Float32

func TestDecodeFloat32__should_decode_float32(t *testing.T) {
	b := []byte{}
	b = EncodeFloat32(b, math.MaxFloat32)

	v, n, err := DecodeFloat32(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, float32(math.MaxFloat32), v)
}

func TestDecodeFloat32__should_decode_float32_from_float64(t *testing.T) {
	b := []byte{}
	b = EncodeFloat64(b, math.MaxFloat32)

	v, n, err := DecodeFloat32(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, float32(math.MaxFloat32), v)
}

// Float64

func TestDecodeFloat64__should_decode_float64(t *testing.T) {
	b := []byte{}
	b = EncodeFloat64(b, math.MaxFloat64)

	v, n, err := DecodeFloat64(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, float64(math.MaxFloat64), v)
}

func TestDecodeFloat64__should_decode_float64_from_float32(t *testing.T) {
	b := []byte{}
	b = EncodeFloat32(b, math.MaxFloat32)

	v, n, err := DecodeFloat64(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, float64(math.MaxFloat32), v)
}

// Bytes

func TestDecodeBytes__should_decode_bytes(t *testing.T) {
	b := []byte{}
	v := []byte("hello, world")

	b, err := EncodeBytes(b, v)
	if err != nil {
		t.Fatal(err)
	}

	v1, n, err := DecodeBytes(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, v, v1)
}

// String

func TestDecodeString__should_decode_string(t *testing.T) {
	b := []byte{}
	v := "hello, world"

	b, err := EncodeString(b, v)
	if err != nil {
		t.Fatal(err)
	}

	v1, n, err := DecodeString(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, v, v1)
}

// List

func TestDecodeList__should_decode_list(t *testing.T) {
	b := testWriteList(t)
	_, n, err := decodeList(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
}

func TestDecodeListTable__should_decode_list_table(t *testing.T) {
	elements := testListElements()

	for i := 0; i <= len(elements); i++ {
		ee0 := elements[i:]

		table0, size, err := encodeListTable(nil, ee0, false)
		if err != nil {
			t.Fatal(err)
		}

		table1, err := decodeListTable(table0, size, false)
		if err != nil {
			t.Fatal(err)
		}

		ee1 := table1.elements(false)
		require.Equal(t, ee0, ee1)
	}
}

func TestDecodeList__should_return_error_when_invalid_type(t *testing.T) {
	b := testWriteList(t)
	b[len(b)-1] = byte(TypeMessage)

	_, n, err := decodeList(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected type")
}

func TestDecodeList__should_return_error_when_invalid_table_size(t *testing.T) {
	b := []byte{}
	b = append(b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff) // varint overflow
	b = append(b, byte(TypeList))

	_, n, err := decodeList(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table size")
}

func TestDecodeList__should_return_error_when_invalid_body_size(t *testing.T) {
	b := []byte{}
	b = append(b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff) // varint overflow
	b = encodeListTableSize(b, 1000)
	b = append(b, byte(TypeList))

	_, n, err := decodeList(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid body size")
}

func TestDecodeList__should_return_error_when_invalid_table(t *testing.T) {
	b, _, err := encodeListTable(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	b = encodeListBodySize(b, 0)
	b = encodeListTableSize(b, 1000)
	b = append(b, byte(TypeList))

	_, n, err := decodeList(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table")
}

func TestDecodeList__should_return_error_when_invalid_body(t *testing.T) {
	b, _, err := encodeListTable(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	b = encodeListBodySize(b, 1000)
	b = encodeListTableSize(b, 0)
	b = append(b, byte(TypeList))

	_, n, err := decodeList(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid body")
}

// Message

func TestDecodeMessage__should_decode_message(t *testing.T) {
	b := testEncodeMessage(t)
	_, n, err := decodeMessage(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
}

func TestDecodeMessageTable__should_decode_message_table(t *testing.T) {
	fields := testMessageFields()

	for i := 0; i <= len(fields); i++ {
		fields0 := fields[i:]

		data, size, err := encodeMessageTable(nil, fields0, false)
		if err != nil {
			t.Fatal(err)
		}

		table1, err := decodeMessageTable(data, size, false)
		if err != nil {
			t.Fatal(err)
		}

		fields1 := table1.fields(false)
		require.Equal(t, fields0, fields1)
	}
}

func TestDecodeMessage__should_return_error_when_invalid_type(t *testing.T) {
	b := testEncodeMessage(t)
	b[len(b)-1] = byte(TypeList)

	_, n, err := decodeMessage(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected type")
}

func TestDecodeMessage__should_return_error_when_invalid_table_size(t *testing.T) {
	b := []byte{}
	b = append(b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff) // varint overflow
	b = append(b, byte(TypeMessage))

	_, n, err := decodeMessage(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table size")
}

func TestDecodeMessage__should_return_error_when_invalid_body_size(t *testing.T) {
	b := []byte{}
	b = append(b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff) // varint overflow
	b = encodeListTableSize(b, 1000)
	b = append(b, byte(TypeMessage))

	_, n, err := decodeMessage(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid body size")
}

func TestDecodeMessage__should_return_error_when_invalid_table(t *testing.T) {
	b, _, err := encodeMessageTable(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	b = encodeListBodySize(b, 0)
	b = encodeListTableSize(b, 1000)
	b = append(b, byte(TypeMessage))

	_, n, err := decodeMessage(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table")
}

func TestDecodeMessage__should_return_error_when_invalid_body(t *testing.T) {
	b, _, err := encodeMessageTable(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	b = encodeListBodySize(b, 1000)
	b = encodeListTableSize(b, 0)
	b = append(b, byte(TypeMessage))

	_, n, err := decodeMessage(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid body")
}
