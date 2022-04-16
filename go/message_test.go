package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testEncodeMessage(t *testing.T) []byte {
	w := NewEncoder()
	w.BeginMessage()

	w.Byte(1)
	w.Field(1)

	w.Int32(1)
	w.Field(3)
	w.Int64(1)
	w.Field(4)

	w.Uint32(1)
	w.Field(12)
	w.Uint64(1)
	w.Field(13)

	w.Float32(1)
	w.Field(20)
	w.Float64(1)
	w.Field(21)

	w.String("hello, world")
	w.Field(30)
	w.Bytes([]byte("hello, world"))
	w.Field(31)

	w.BeginList()
	w.String("element1")
	w.Element()
	w.End()
	w.Field(40)

	w.BeginMessage()
	w.String("field1")
	w.Field(1)
	w.End()
	w.Field(50)

	b, err := w.End()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestMessage_Getters__should_access_fields(t *testing.T) {
	b := testEncodeMessage(t)
	m, _, err := DecodeMessage(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, byte(1), m.GetByte(1))

	assert.Equal(t, int32(1), m.GetInt32(3))
	assert.Equal(t, int64(1), m.GetInt64(4))

	assert.Equal(t, uint32(1), m.GetUint32(12))
	assert.Equal(t, uint64(1), m.GetUint64(13))

	assert.Equal(t, float32(1), m.GetFloat32(20))
	assert.Equal(t, float64(1), m.GetFloat64(21))

	assert.Equal(t, "hello, world", m.GetString(30))
	assert.Equal(t, []byte("hello, world"), m.GetBytes(31))

	assert.Equal(t, 1, m.GetMessage(50).Len())
}
