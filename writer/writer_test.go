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

	assert.Equal(t, writerClosed, w.err)
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
