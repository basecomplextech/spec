package writer

import (
	"testing"
	"unsafe"

	"github.com/complex1tech/spec/go/encoding"
	"github.com/stretchr/testify/assert"
)

// func TestWriter__should_write_message(t *testing.T) {
// 	msg := newTestObject()

// 	w := NewWriter()
// 	b, err := BuildTestMessageWriter(w)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := msg.Encode(b); err != nil {
// 		t.Fatal(err)
// 	}
// 	data, err := w.End()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	msg1 := &TestObject{}
// 	if err := msg1.Decode(data); err != nil {
// 		t.Fatal(err)
// 	}

// 	assert.Equal(t, msg, msg1)
// }

// func TestWriter__should_close_on_end(t *testing.T) {
// 	msg := newTestSmall()

// 	w := NewWriter()
// 	_, err := msg.Encode(w)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	assert.Equal(t, writerClosed, w.err)
// 	assert.Nil(t, w.writerState)
// }

// List

// func TestWriter__should_write_list(t *testing.T) {
// 	w := NewWriter()
// 	w.BeginList()

// 	w.Int64(1)
// 	w.Element()

// 	w.Int64(2)
// 	w.Element()

// 	w.Int64(3)
// 	w.Element()

// 	b, err := w.End()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	list, _, err := DecodeList(b, DecodeInt64)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	items1 := []int64{
// 		list.Get(0),
// 		list.Get(1),
// 		list.Get(2),
// 	}

// 	assert.Equal(t, []int64{1, 2, 3}, items1)
// }

// End

func TestWriter_End__should_return_finished_bytes(t *testing.T) {
	w := NewWriter()
	w.BeginMessage()

	b, err := w.End()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, byte(encoding.TypeMessage), b[len(b)-1])
}

func TestWriter_End__should_return_error_when_not_finished(t *testing.T) {
	w := NewWriter()
	w.Bool(true)

	_, err := w.End()
	assert.Error(t, err)
}

// Data

func TestWriter_Data__should_return_error_when_unconsumed_data(t *testing.T) {
	w := NewWriter()
	w.BeginMessage()
	w.Int64(1)

	err := w.Int64(1)
	assert.Error(t, err)
}

// Element

func TestWriter_Element__should_return_error_when_not_in_list(t *testing.T) {
	w := NewWriter()
	err := w.Element()
	assert.Error(t, err)
}

func TestWriter_Element__should_return_error_when_in_message(t *testing.T) {
	w := NewWriter()
	w.BeginMessage()

	err := w.Element()
	assert.Error(t, err)
}

// Field

func TestWriter_Field__should_return_error_when_not_in_message(t *testing.T) {
	w := NewWriter()
	err := w.Field(1)
	assert.Error(t, err)
}

func TestWriter_Field__should_return_error_when_in_list(t *testing.T) {
	w := NewWriter()
	w.BeginList()
	err := w.Field(1)
	assert.Error(t, err)
}

// Struct size

func TestWriter__struct_size_must_be_less_or_equal_2048(t *testing.T) {
	// 2048 is 1/2 of 4kb page or 1/4 of 8kb page.

	w := Writer{}
	size := unsafe.Sizeof(w)

	assert.LessOrEqual(t, int(size), 2048)
}
