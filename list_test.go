package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testWriteList(t *testing.T) []byte {
	w := NewWriter()
	w.BeginList()

	w.Int8(1)
	w.Element()
	w.Int16(1)
	w.Element()
	w.Int32(1)
	w.Element()
	w.Int64(1)
	w.Element()

	w.UInt8(1)
	w.Element()
	w.UInt16(1)
	w.Element()
	w.UInt32(1)
	w.Element()
	w.UInt64(1)
	w.Element()

	w.Float32(1)
	w.Element()
	w.Float64(1)
	w.Element()

	w.String("hello, world")
	w.Element()
	w.Bytes([]byte("hello, world"))
	w.Element()

	w.BeginList()
	w.String("element1")
	w.Element()
	w.EndList()
	w.Element()

	w.BeginMessage()
	w.String("field1")
	w.Field(1)
	w.EndMessage()
	w.Element()

	w.EndList()
	b, err := w.End()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// Getters

func TestList_Getters__should_access_elements(t *testing.T) {
	b := testWriteList(t)
	l, err := GetList(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int8(1), l.Int8(0))
	assert.Equal(t, int16(1), l.Int16(1))
	assert.Equal(t, int32(1), l.Int32(2))
	assert.Equal(t, int64(1), l.Int64(3))

	assert.Equal(t, uint8(1), l.UInt8(4))
	assert.Equal(t, uint16(1), l.UInt16(5))
	assert.Equal(t, uint32(1), l.UInt32(6))
	assert.Equal(t, uint64(1), l.UInt64(7))

	assert.Equal(t, float32(1), l.Float32(8))
	assert.Equal(t, float64(1), l.Float64(9))

	assert.Equal(t, "hello, world", l.String(10))
	assert.Equal(t, []byte("hello, world"), l.Bytes(11))

	assert.Equal(t, 1, l.List(12).Len())
	assert.Equal(t, 1, l.Message(13).Len())
}
