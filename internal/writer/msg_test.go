package writer

import (
	"math"
	"testing"

	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/spec/encoding"
	"github.com/stretchr/testify/assert"
)

func TestMessageWriter__should_write_message(t *testing.T) {
	w := testWriter()

	w1 := w.Message()
	w1.Field(1).Bool(true)
	w1.Field(2).Byte(2)

	w1.Field(10).Int32(math.MaxInt32)
	w1.Field(11).Int64(math.MaxInt64)

	w1.Field(20).Uint32(math.MaxUint32)
	w1.Field(21).Uint64(math.MaxUint64)

	w1.Field(30).Float32(math.MaxFloat32)
	w1.Field(31).Float64(math.MaxFloat64)

	w1.Field(40).Bin64(bin.Random64())
	w1.Field(41).Bin128(bin.Random128())
	w1.Field(42).Bin256(bin.Random256())

	w1.Field(50).String("hello world")
	w1.Field(51).Bytes([]byte("hello world"))

	list1 := w1.Field(60).List()
	list1.String("sublist")
	if err := list1.End(); err != nil {
		t.Fatal(err)
	}

	msg1 := w1.Field(61).Message()
	msg1.Field(1).String("submessage")
	if err := msg1.End(); err != nil {
		t.Fatal(err)
	}

	b, err := w1.Build()
	if err != nil {
		t.Fatal(err)
	}

	_, n, err := encoding.DecodeMessageMeta(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(b), n)
}
