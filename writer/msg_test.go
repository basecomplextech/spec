package writer

import (
	"math"
	"testing"

	"github.com/complex1tech/baselibrary/bin"
	"github.com/complex1tech/spec/encoding"
	"github.com/stretchr/testify/assert"
)

func TestMessageWriter__should_write_message(t *testing.T) {
	w := testWriter()

	msg := w.Message()
	msg.Field(1).Bool(true)
	msg.Field(2).Byte(2)

	msg.Field(10).Int32(math.MaxInt32)
	msg.Field(11).Int64(math.MaxInt64)

	msg.Field(20).Uint32(math.MaxUint32)
	msg.Field(21).Uint64(math.MaxUint64)

	msg.Field(30).Float32(math.MaxFloat32)
	msg.Field(31).Float64(math.MaxFloat64)

	msg.Field(40).Bin64(bin.Random64())
	msg.Field(41).Bin128(bin.Random128())
	msg.Field(42).Bin256(bin.Random256())

	msg.Field(50).String("hello world")
	msg.Field(51).Bytes([]byte("hello world"))

	list1 := msg.Field(60).List()
	list1.String("sublist")
	if err := list1.End(); err != nil {
		t.Fatal(err)
	}

	msg1 := msg.Field(61).Message()
	msg1.Field(1).String("submessage")
	if err := msg1.End(); err != nil {
		t.Fatal(err)
	}

	b, err := msg.Build()
	if err != nil {
		t.Fatal(err)
	}

	_, n, err := encoding.DecodeMessageMeta(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(b), n)
}
