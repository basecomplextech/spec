package spec

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ReadType

func TestReadType__should_return_type(t *testing.T) {
	b := []byte{}
	b = append(b, byte(TypeString))

	v, err := ReadType(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, v, TypeString)
}

func TestReadType__should_return_nil_when_empty(t *testing.T) {
	b := []byte{}

	v, err := ReadType(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, v, TypeNil)
}

// ReadBool

func TestReadBool__should_read_bool_value(t *testing.T) {
	v, err := ReadBool([]byte{byte(TypeTrue)})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, v)

	v, err = ReadBool([]byte{byte(TypeFalse)})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, false, v)
}

// ReadInt8

func TestReadInt8__should_read_int8(t *testing.T) {
	b := []byte{}
	b = writeInt8(b, 1)

	v, err := ReadInt8(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int8(1), v)
}

func TestReadInt8__should_read_int8_from_int16(t *testing.T) {
	b := []byte{}
	b = writeInt16(b, 1)

	v, err := ReadInt8(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int8(1), v)
}

func TestReadInt8__should_read_int8_from_int32(t *testing.T) {
	b := []byte{}
	b = writeInt32(b, 1)

	v, err := ReadInt8(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int8(1), v)
}

func TestReadInt8__should_read_int8_from_int64(t *testing.T) {
	b := []byte{}
	b = writeInt64(b, 1)

	v, err := ReadInt8(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int8(1), v)
}

// Int16

func TestReadInt16__should_read_int16(t *testing.T) {
	b := []byte{}
	b = writeInt16(b, 1)

	v, err := ReadInt16(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int16(1), v)
}

// Int32

func TestReadInt32__should_read_int32(t *testing.T) {
	b := []byte{}
	b = writeInt32(b, 1)

	v, err := ReadInt32(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int32(1), v)
}

// Int64

func TestReadInt64__should_read_int64(t *testing.T) {
	b := []byte{}
	b = writeInt64(b, 1)

	v, err := ReadInt64(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(1), v)
}

func TestReadInt64__should_read_int64_from_int8(t *testing.T) {
	b := []byte{}
	b = writeInt8(b, 1)

	v, err := ReadInt64(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(1), v)
}

func TestReadInt64__should_read_int64_from_int16(t *testing.T) {
	b := []byte{}
	b = writeInt16(b, 1)

	v, err := ReadInt64(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(1), v)
}

func TestReadInt64__should_read_int64_from_int32(t *testing.T) {
	b := []byte{}
	b = writeInt32(b, 1)

	v, err := ReadInt64(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(1), v)
}

// Float32

func TestReadFloat32__should_read_float32(t *testing.T) {
	b := []byte{}
	b = writeFloat32(b, math.MaxFloat32)

	v, err := ReadFloat32(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, float32(math.MaxFloat32), v)
}

func TestReadFloat32__should_read_float32_from_float64(t *testing.T) {
	b := []byte{}
	b = writeFloat64(b, math.MaxFloat32)

	v, err := ReadFloat32(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, float32(math.MaxFloat32), v)
}

// Float64

func TestReadFloat64__should_read_float64(t *testing.T) {
	b := []byte{}
	b = writeFloat64(b, math.MaxFloat64)

	v, err := ReadFloat64(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, float64(math.MaxFloat64), v)
}

func TestReadFloat64__should_read_float64_from_float32(t *testing.T) {
	b := []byte{}
	b = writeFloat32(b, math.MaxFloat32)

	v, err := ReadFloat64(b)
	if err != nil {
		t.Fatal(err)
	}
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

	v1, err := ReadBytes(b)
	if err != nil {
		t.Fatal(err)
	}

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

	v1, err := ReadString(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, v, v1)
}

// List

func TestReadList__should_read_list(t *testing.T) {
	b := testWriteList(t)
	_, err := ReadList(b)
	if err != nil {
		t.Fatal(err)
	}
}

// List table

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

// Message

func TestReadMessage__should_read_message(t *testing.T) {
	b := testWriteMessage(t)
	_, err := ReadMessage(b)
	if err != nil {
		t.Fatal(err)
	}
}
