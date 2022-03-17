package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testWriteMessage(t *testing.T) []byte {
	w := NewWriter()
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
	w.EndList()
	w.Field(40)

	w.BeginMessage()
	w.String("field1")
	w.Field(1)
	w.EndMessage()
	w.Field(50)

	w.EndMessage()
	b, err := w.End()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestMessage_Getters__should_access_fields(t *testing.T) {
	b := testWriteMessage(t)
	m, _, err := ReadMessage(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, byte(1), m.Byte(1))

	assert.Equal(t, int32(1), m.Int32(3))
	assert.Equal(t, int64(1), m.Int64(4))

	assert.Equal(t, uint32(1), m.Uint32(12))
	assert.Equal(t, uint64(1), m.Uint64(13))

	assert.Equal(t, float32(1), m.Float32(20))
	assert.Equal(t, float64(1), m.Float64(21))

	assert.Equal(t, "hello, world", m.String(30))
	assert.Equal(t, []byte("hello, world"), m.Bytes(31))

	assert.Equal(t, 1, m.Message(50).Count())
}
