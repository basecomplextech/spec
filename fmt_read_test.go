package spec

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readType

func TestReadType__should_return_type(t *testing.T) {
	b := []byte{}
	b = append(b, byte(TypeString))

	v, n, err := readType(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, v, TypeString)
}

func TestReadType__should_return_nil_when_empty(t *testing.T) {
	b := []byte{}

	v, n, err := readType(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, v, TypeNil)
}

// readBool

func TestReadBool__should_read_bool_value(t *testing.T) {
	b := []byte{byte(TypeTrue)}
	v, n, err := readBool(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, true, v)

	b = []byte{byte(TypeFalse)}
	v, n, err = readBool(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, false, v)
}

// readInt8

func TestReadInt8__should_read_int8(t *testing.T) {
	b := []byte{}
	b = writeInt8(b, 1)

	v, n, err := readInt8(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, int8(1), v)
}

func TestReadInt8__should_read_int8_from_int16(t *testing.T) {
	b := []byte{}
	b = writeInt16(b, 1)

	v, n, err := readInt8(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, int8(1), v)
}

func TestReadInt8__should_read_int8_from_int32(t *testing.T) {
	b := []byte{}
	b = writeInt32(b, 1)

	v, n, err := readInt8(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, int8(1), v)
}

func TestReadInt8__should_read_int8_from_int64(t *testing.T) {
	b := []byte{}
	b = writeInt64(b, 1)

	v, n, err := readInt8(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, int8(1), v)
}

// Int16

func TestReadInt16__should_read_int16(t *testing.T) {
	b := []byte{}
	b = writeInt16(b, 1)

	v, n, err := readInt16(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, int16(1), v)
}

// Int32

func TestReadInt32__should_read_int32(t *testing.T) {
	b := []byte{}
	b = writeInt32(b, 1)

	v, n, err := readInt32(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, int32(1), v)
}

// Int64

func TestReadInt64__should_read_int64(t *testing.T) {
	b := []byte{}
	b = writeInt64(b, 1)

	v, n, err := readInt64(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, int64(1), v)
}

func TestReadInt64__should_read_int64_from_int8(t *testing.T) {
	b := []byte{}
	b = writeInt8(b, 1)

	v, n, err := readInt64(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, int64(1), v)
}

func TestReadInt64__should_read_int64_from_int16(t *testing.T) {
	b := []byte{}
	b = writeInt16(b, 1)

	v, n, err := readInt64(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, int64(1), v)
}

func TestReadInt64__should_read_int64_from_int32(t *testing.T) {
	b := []byte{}
	b = writeInt32(b, 1)

	v, n, err := readInt64(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, int64(1), v)
}

// Float32

func TestReadFloat32__should_read_float32(t *testing.T) {
	b := []byte{}
	b = writeFloat32(b, math.MaxFloat32)

	v, n, err := readFloat32(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, float32(math.MaxFloat32), v)
}

func TestReadFloat32__should_read_float32_from_float64(t *testing.T) {
	b := []byte{}
	b = writeFloat64(b, math.MaxFloat32)

	v, n, err := readFloat32(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, float32(math.MaxFloat32), v)
}

// Float64

func TestReadFloat64__should_read_float64(t *testing.T) {
	b := []byte{}
	b = writeFloat64(b, math.MaxFloat64)

	v, n, err := readFloat64(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, float64(math.MaxFloat64), v)
}

func TestReadFloat64__should_read_float64_from_float32(t *testing.T) {
	b := []byte{}
	b = writeFloat32(b, math.MaxFloat32)

	v, n, err := readFloat64(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, float64(math.MaxFloat32), v)
}

// Bytes

func TestReadBytes__should_read_bytes(t *testing.T) {
	b := []byte{}
	v := []byte("hello, world")

	b, err := writeBytes(b, v)
	if err != nil {
		t.Fatal(err)
	}

	v1, n, err := readBytes(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, v, v1)
}

// String

func TestReadString__should_read_string(t *testing.T) {
	b := []byte{}
	v := "hello, world"

	b, err := writeString(b, v)
	if err != nil {
		t.Fatal(err)
	}

	v1, n, err := readString(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, v, v1)
}

// List

func TestReadList__should_read_list(t *testing.T) {
	b := testWriteList(t)
	_, n, err := readList(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
}

func TestReadListTable__should_read_list_table(t *testing.T) {
	elements := testListElements()

	for i := 0; i <= len(elements); i++ {
		ee0 := elements[i:]

		table0, size, err := _writeListTable(nil, ee0, false)
		if err != nil {
			t.Fatal(err)
		}

		table1, err := _readListTable(table0, size, false)
		if err != nil {
			t.Fatal(err)
		}

		ee1 := table1.elements(false)
		require.Equal(t, ee0, ee1)
	}
}

func TestReadList__should_return_error_when_invalid_type(t *testing.T) {
	b := testWriteList(t)
	b[len(b)-1] = byte(TypeMessage)

	_, n, err := readList(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected type")
}

func TestReadList__should_return_error_when_invalid_table_size(t *testing.T) {
	b := []byte{}
	b = append(b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff) // varint overflow
	b = append(b, byte(TypeList))

	_, n, err := readList(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table size")
}

func TestReadList__should_return_error_when_invalid_body_size(t *testing.T) {
	b := []byte{}
	b = append(b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff) // varint overflow
	b = _writeListTableSize(b, 1000)
	b = append(b, byte(TypeList))

	_, n, err := readList(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid body size")
}

func TestReadList__should_return_error_when_invalid_table(t *testing.T) {
	b, _, err := _writeListTable(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	b = _writeListBodySize(b, 0)
	b = _writeListTableSize(b, 1000)
	b = append(b, byte(TypeList))

	_, n, err := readList(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table")
}

func TestReadList__should_return_error_when_invalid_body(t *testing.T) {
	b, _, err := _writeListTable(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	b = _writeListBodySize(b, 1000)
	b = _writeListTableSize(b, 0)
	b = append(b, byte(TypeList))

	_, n, err := readList(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid body")
}

// Message

func TestReadMessage__should_read_message(t *testing.T) {
	b := testWriteMessage(t)
	_, n, err := readMessage(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
}

func TestReadMessageTable__should_read_message_table(t *testing.T) {
	fields := testMessageFields()

	for i := 0; i <= len(fields); i++ {
		fields0 := fields[i:]

		data, size, err := _writeMessageTable(nil, fields0, false)
		if err != nil {
			t.Fatal(err)
		}

		table1, err := _readMessageTable(data, size, false)
		if err != nil {
			t.Fatal(err)
		}

		fields1 := table1.fields(false)
		require.Equal(t, fields0, fields1)
	}
}

func TestReadMessage__should_return_error_when_invalid_type(t *testing.T) {
	b := testWriteMessage(t)
	b[len(b)-1] = byte(TypeList)

	_, n, err := readMessage(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected type")
}

func TestReadMessage__should_return_error_when_invalid_table_size(t *testing.T) {
	b := []byte{}
	b = append(b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff) // varint overflow
	b = append(b, byte(TypeMessage))

	_, n, err := readMessage(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table size")
}

func TestReadMessage__should_return_error_when_invalid_body_size(t *testing.T) {
	b := []byte{}
	b = append(b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff) // varint overflow
	b = _writeListTableSize(b, 1000)
	b = append(b, byte(TypeMessage))

	_, n, err := readMessage(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid body size")
}

func TestReadMessage__should_return_error_when_invalid_table(t *testing.T) {
	b, _, err := _writeMessageTable(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	b = _writeListBodySize(b, 0)
	b = _writeListTableSize(b, 1000)
	b = append(b, byte(TypeMessage))

	_, n, err := readMessage(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table")
}

func TestReadMessage__should_return_error_when_invalid_body(t *testing.T) {
	b, _, err := _writeMessageTable(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	b = _writeListBodySize(b, 1000)
	b = _writeListTableSize(b, 0)
	b = append(b, byte(TypeMessage))

	_, n, err := readMessage(b)
	assert.Equal(t, -1, n)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid body")
}
