package spec

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuffer_Alloc__should_allocate_scratch_slice(t *testing.T) {
	buf := newBuffer(nil)

	b := buf.Alloc(8)
	binary.BigEndian.PutUint64(b, math.MaxUint64)

	n, err := buf.Write(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 8, n)

	bytes := buf.Bytes()
	i := binary.BigEndian.Uint64(bytes)
	assert.Equal(t, uint64(math.MaxUint64), i)
}

func TestBuffer_Write__should_write_bytes(t *testing.T) {
	buf := newBuffer(nil)

	n, err := buf.Write([]byte("hello, "))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len("hello, "), n)

	n, err = buf.Write([]byte("world"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len("world"), n)

	bytes := buf.Bytes()
	assert.Equal(t, []byte("hello, world"), bytes)
}
