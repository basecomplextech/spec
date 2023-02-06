package writer

import (
	"testing"
	"unsafe"

	"github.com/complex1tech/baselibrary/buffer"
	"github.com/complex1tech/spec/encoding"
	"github.com/stretchr/testify/assert"
)

func testWriter() *writer {
	buf := buffer.New()
	return newWriter(buf)
}

// Reset

func TestWriter_Reset__should_reset_writer(t *testing.T) {
	w := testWriter()
	w.closef("test error")

	buf := buffer.New()
	w.Reset(buf)

	assert.NotNil(t, w.writerState)
	assert.Nil(t, w.err)
}

func TestWriter_Reset__should_create_buffer_when_nil(t *testing.T) {
	w := testWriter()
	w.closef("test error")

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

	assert.Equal(t, byte(encoding.TypeMessage), b[len(b)-1])
}

func TestWriter_end__should_return_error_when_not_finished(t *testing.T) {
	w := testWriter()
	w.Value().Bool(true)

	_, err := w.end()
	assert.Error(t, err)
}

func TestWriter_end__should_close_writer_on_root_element_end(t *testing.T) {
	w := testWriter()

	msg := w.Message()
	msg.Field(1).Bool(true)
	msg.Field(2).Byte(2)

	_, err := msg.Build()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, errClosed, w.err)
	assert.Nil(t, w.writerState)
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
