package writer

import (
	"testing"
	"unsafe"

	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testWriter() *writer {
	return newWriter(nil, false /* auto release */)
}

// Reset

func TestWriter_Reset__should_reset_writer(t *testing.T) {
	w := testWriter()
	w.failf("test error")

	buf := buffer.New()
	w.Reset(buf)

	assert.NotNil(t, w.writerState)
	assert.Nil(t, w.err)
}

func TestWriter_Reset__should_create_buffer_when_nil(t *testing.T) {
	w := testWriter()
	w.failf("test error")

	w.Reset(nil)

	assert.NotNil(t, w.writerState)
	assert.Nil(t, w.err)
}

func TestWriter_Reset__should_reset_existing_state(t *testing.T) {
	w := testWriter()
	w.Value().String("hello, world")

	s := w.writerState
	w.Reset(nil)

	s1 := w.writerState
	assert.Same(t, s, s1)
}

// Free

func TestWriter_Free__should_close_writer_and_release_state(t *testing.T) {
	w := testWriter()
	w.Value().String("hello, world")
	w.Free()

	assert.Equal(t, errClosed, w.err)
	assert.Nil(t, w.writerState)
}

// end

func TestWriter_end__should_return_finished_bytes(t *testing.T) {
	w := testWriter()
	w.beginMessage()

	b, err := w.end()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, byte(core.TypeMessage), b[len(b)-1])
}

func TestWriter_end__should_return_error_when_no_data(t *testing.T) {
	w := testWriter()

	_, err := w.end()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stack is empty")
}

func TestWriter_end__should_return_error_when_not_root_value(t *testing.T) {
	w := testWriter()
	w.beginMessage()
	w.beginField(1)
	w.Value().Bool(true)

	_, err := w.end()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not root value")
}

func TestWriter_end__should_close_writer_on_root_end(t *testing.T) {
	w := testWriter()

	msg := w.Message()
	msg.Field(1).Bool(true)
	msg.Field(2).Byte(2)

	_, err := msg.Build()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, errClosed, w.err)
}

func TestWriter_end__should_free_when_autoreleasing(t *testing.T) {
	w := newWriter(nil, true /* auto release */)

	msg := w.Message()
	msg.Field(1).Bool(true)
	msg.Field(2).Byte(2)

	_, err := msg.Build()
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, w.writerState)
}

func TestWriter_end__should_not_free_when_not_autoreleasing(t *testing.T) {
	w := newWriter(nil, false /* no auto release */)

	msg := w.Message()
	msg.Field(1).Bool(true)
	msg.Field(2).Byte(2)

	_, err := msg.Build()
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, w.writerState)
}

// pushData

func TestWriter_pushData__should_return_error_when_unconsumed_data(t *testing.T) {
	w := testWriter()
	w.beginMessage()
	w.Value().Int64(1)

	err := w.Value().Int64(1)
	assert.Error(t, err)
}

// element

func TestWriter_element__should_return_error_when_not_list(t *testing.T) {
	w := testWriter()
	err := w.element()
	assert.Error(t, err)
}

func TestWriter_element__should_return_error_when_in_message(t *testing.T) {
	w := testWriter()
	w.beginMessage()

	err := w.element()
	assert.Error(t, err)
}

// Field

func TestWriter_Field__should_return_error_when_not_in_message(t *testing.T) {
	w := testWriter()
	err := w.field(1)
	assert.Error(t, err)
}

func TestWriter_Field__should_return_error_when_in_list(t *testing.T) {
	w := testWriter()
	w.beginList()
	err := w.field(1)
	assert.Error(t, err)
}

// Struct size

func TestWriter__struct_size_must_be_less_or_equal_2048(t *testing.T) {
	// 2048 is 1/2 of 4kb page or 1/4 of 8kb page.

	w := writer{}
	size := unsafe.Sizeof(w)

	assert.LessOrEqual(t, int(size), 2048)
}
