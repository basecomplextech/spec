package spec

import (
	"testing"
	"unsafe"

	"github.com/complex1tech/baselibrary/buffer"
	"github.com/stretchr/testify/assert"
)

func TestWriter(t *testing.T) {
	o := newTestObject()

	var data0 []byte
	{
		e := newWriter(buffer.New())
		b, err := BuildTestMessageWriter(e)
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Encode(b); err != nil {
			t.Fatal(err)
		}

		data0, err = e.End()
		if err != nil {
			t.Fatal(err)
		}
	}

	var data1 []byte
	{
		buf := buffer.New()
		e := newWriter(buf)

		var err error
		data1, err = o.Write(e)
		if err != nil {
			t.Fatal(err)
		}
	}

	// assert.Equal(t, data0, data1)

	o0 := &TestObject{}
	if err := o0.Decode(data0); err != nil {
		t.Fatal(err)
	}

	o1 := &TestObject{}
	if err := o1.Decode(data1); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, o0, o1)
}

func TestWriter__should_write_message(t *testing.T) {
	msg := newTestObject()

	e := NewWriter()
	b, err := BuildTestMessageWriter(e)
	if err != nil {
		t.Fatal(err)
	}
	if err := msg.Encode(b); err != nil {
		t.Fatal(err)
	}
	data, err := e.End()
	if err != nil {
		t.Fatal(err)
	}

	msg1 := &TestObject{}
	if err := msg1.Decode(data); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, msg, msg1)
}

func TestWriter__should_close_on_end(t *testing.T) {
	msg := newTestSmall()

	e := NewWriter()
	_, err := msg.Encode(e)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, writerClosed, e.err)
	assert.Nil(t, e.writerState)
}

// List

func TestWriter__should_write_list(t *testing.T) {
	e := NewWriter()
	e.BeginList()

	e.Int64(1)
	e.Element()

	e.Int64(2)
	e.Element()

	e.Int64(3)
	e.Element()

	b, err := e.End()
	if err != nil {
		t.Fatal(err)
	}

	list, _, err := DecodeList(b, DecodeInt64)
	if err != nil {
		t.Fatal(err)
	}
	items1 := []int64{
		list.Get(0),
		list.Get(1),
		list.Get(2),
	}

	assert.Equal(t, []int64{1, 2, 3}, items1)
}

// End

func TestWriter_End__should_return_finished_bytes(t *testing.T) {
	e := NewWriter()
	e.BeginMessage()

	b, err := e.End()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, byte(TypeMessage), b[len(b)-1])
}

func TestWriter_End__should_return_error_when_not_finished(t *testing.T) {
	e := NewWriter()
	e.Bool(true)

	_, err := e.End()
	assert.Error(t, err)
}

// Data

func TestWriter_Data__should_return_error_when_unconsumed_data(t *testing.T) {
	e := NewWriter()
	e.BeginMessage()
	e.Int64(1)

	err := e.Int64(1)
	assert.Error(t, err)
}

// Element

func TestWriter_Element__should_return_error_when_not_in_list(t *testing.T) {
	e := NewWriter()
	err := e.Element()
	assert.Error(t, err)
}

func TestWriter_Element__should_return_error_when_in_message(t *testing.T) {
	e := NewWriter()
	e.BeginMessage()

	err := e.Element()
	assert.Error(t, err)
}

// Field

func TestWriter_Field__should_return_error_when_not_in_message(t *testing.T) {
	e := NewWriter()
	err := e.Field(1)
	assert.Error(t, err)
}

func TestWriter_Field__should_return_error_when_in_list(t *testing.T) {
	e := NewWriter()
	e.BeginList()
	err := e.Field(1)
	assert.Error(t, err)
}

// Struct size

func TestWriter__struct_size_must_be_less_or_equal_2048(t *testing.T) {
	// 2048 is 1/2 of 4kb page or 1/4 of 8kb page.

	e := Writer{}
	size := unsafe.Sizeof(e)

	assert.LessOrEqual(t, int(size), 2048)
}
