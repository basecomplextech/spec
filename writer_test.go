package spec

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestWriter_size__must_be_less_or_equal_2048(t *testing.T) {
	// 2048 is 1/2 of 4kb page or 1/4 of 8kb page.

	w := Writer{}
	size := unsafe.Sizeof(w)

	assert.LessOrEqual(t, int(size), 2048)
}

func TestWriter_Write(t *testing.T) {
	msg := newTestMessage()

	w := NewWriter()
	if err := msg.Write(w); err != nil {
		t.Fatal(err)
	}
	b, err := w.End()
	if err != nil {
		t.Fatal(err)
	}

	msg1 := &TestMessage{}
	if err := msg1.Read(ReadData(b)); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, msg, msg1)
}
