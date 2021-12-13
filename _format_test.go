package protocol

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriterReader(t *testing.T) {
	w := NewWriter()
	w.Bool(true)
	w.Byte(1)

	w.Int8(math.MaxInt8)
	w.Int16(math.MaxInt16)
	w.Int32(math.MaxInt32)
	w.Int64(math.MaxInt64)

	w.UInt8(math.MaxUint8)
	w.UInt16(math.MaxUint16)
	w.UInt32(math.MaxUint32)
	w.UInt64(math.MaxUint64)

	w.Float32(math.MaxFloat32)
	w.Float64(math.MaxFloat64)

	w.Bytes([]byte("hello, world"))
	w.String("hello, world")
	data := w.End()

	r := NewReader(data)
	assert.Equal(t, true, r.Bool())
	assert.Equal(t, byte(1), r.Byte())

	assert.Equal(t, int8(math.MaxInt8), r.Int8())
	assert.Equal(t, int16(math.MaxInt16), r.Int16())
	assert.Equal(t, int32(math.MaxInt32), r.Int32())
	assert.Equal(t, int64(math.MaxInt64), r.Int64())

	assert.Equal(t, uint8(math.MaxUint8), r.UInt8())
	assert.Equal(t, uint16(math.MaxUint16), r.UInt16())
	assert.Equal(t, uint32(math.MaxUint32), r.UInt32())
	assert.Equal(t, uint64(math.MaxUint64), r.Uint64())

	assert.Equal(t, float32(math.MaxFloat32), r.Float32())
	assert.Equal(t, float64(math.MaxFloat64), r.Float64())

	assert.Equal(t, []byte("hello, world"), r.Bytes())
	assert.Equal(t, "hello, world", r.String())
}
