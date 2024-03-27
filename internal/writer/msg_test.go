package writer

import (
	"math"
	"testing"

	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/spec/encoding"
	"github.com/basecomplextech/spec/internal/types"
	"github.com/stretchr/testify/assert"
)

func testWriteMessage(t *testing.T, w MessageWriter) {
	w.Field(1).Bool(true)
	w.Field(2).Byte(2)

	w.Field(10).Int32(math.MaxInt32)
	w.Field(11).Int64(math.MaxInt64)

	w.Field(20).Uint32(math.MaxUint32)
	w.Field(21).Uint64(math.MaxUint64)

	w.Field(30).Float32(math.MaxFloat32)
	w.Field(31).Float64(math.MaxFloat64)

	w.Field(40).Bin64(bin.Random64())
	w.Field(41).Bin128(bin.Random128())
	w.Field(42).Bin256(bin.Random256())

	w.Field(50).String("hello world")
	w.Field(51).Bytes([]byte("hello world"))

	list1 := w.Field(60).List()
	list1.String("sublist")
	if err := list1.End(); err != nil {
		t.Fatal(err)
	}

	msg1 := w.Field(61).Message()
	msg1.Field(1).String("submessage")
	if err := msg1.End(); err != nil {
		t.Fatal(err)
	}
}

func TestMessageWriter__should_write_message(t *testing.T) {
	w := testWriter()
	w1 := w.Message()
	testWriteMessage(t, w1)

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

func TestMessageWriter_Copy__should_copy_message(t *testing.T) {
	var msg types.Message
	{
		w := testWriter()
		w1 := w.Message()
		testWriteMessage(t, w1)

		b, err := w1.Build()
		if err != nil {
			t.Fatal(err)
		}
		msg, err = types.NewMessageErr(b)
		if err != nil {
			t.Fatal(err)
		}
	}

	w := testWriter()
	w1 := w.Message()
	if err := w1.Copy(msg); err != nil {
		t.Fatal(err)
	}

	b, err := w1.Build()
	if err != nil {
		t.Fatal(err)
	}
	msg1, err := types.NewMessageErr(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, msg.Raw(), msg1.Raw())
}
