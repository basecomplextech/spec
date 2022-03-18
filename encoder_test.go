package spec

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestEncoder__should_encode_message(t *testing.T) {
	msg := newTestObject()

	e := NewEncoder()
	me := BeginTestMessage(e)
	if err := msg.Encode(me); err != nil {
		t.Fatal(err)
	}
	b, err := e.End()
	if err != nil {
		t.Fatal(err)
	}

	msg1 := &TestObject{}
	if err := msg1.Decode(b); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, msg, msg1)
}

// List

func TestEncoder__should_encode_list(t *testing.T) {
	e := NewEncoder()
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
		list.Element(0),
		list.Element(1),
		list.Element(2),
	}

	assert.Equal(t, []int64{1, 2, 3}, items1)
}

// End

func TestEncoder_End__should_return_finished_bytes(t *testing.T) {
	e := NewEncoder()
	e.BeginMessage()

	b, err := e.End()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, byte(TypeMessage), b[len(b)-1])
}

func TestEncoder_End__should_return_error_when_not_finished(t *testing.T) {
	e := NewEncoder()
	e.Bool(true)

	_, err := e.End()
	assert.Error(t, err)
}

// Data

func TestEncoder_Data__should_return_error_when_unconsumed_data(t *testing.T) {
	e := NewEncoder()
	e.BeginMessage()
	e.Int64(1)

	err := e.Int64(1)
	assert.Error(t, err)
}

// Element

func TestEncoder_Element__should_return_error_when_not_in_list(t *testing.T) {
	e := NewEncoder()
	err := e.Element()
	assert.Error(t, err)
}

func TestEncoder_Element__should_return_error_when_in_message(t *testing.T) {
	e := NewEncoder()
	e.BeginMessage()

	err := e.Element()
	assert.Error(t, err)
}

// Field

func TestEncoder_Field__should_return_error_when_not_in_message(t *testing.T) {
	e := NewEncoder()
	err := e.Field(1)
	assert.Error(t, err)
}

func TestEncoder_Field__should_return_error_when_in_list(t *testing.T) {
	e := NewEncoder()
	e.BeginList()
	err := e.Field(1)
	assert.Error(t, err)
}

// Struct size

func TestEncoder__struct_size_must_be_less_or_equal_2048(t *testing.T) {
	// 2048 is 1/2 of 4kb page or 1/4 of 8kb page.

	e := Encoder{}
	size := unsafe.Sizeof(e)

	assert.LessOrEqual(t, int(size), 2048)
}
