package writer

import (
	"testing"

	"github.com/complex1tech/spec/encoding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueWriter_End__should_return_root_value_bytes(t *testing.T) {
	w := testWriter()

	v := w.Value()
	v.String("hello, world")

	b, err := v.Build()
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := encoding.DecodeString(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "hello, world", s.Unwrap())
}

func TestValueWriter_End__should_return_error_when_not_root_value(t *testing.T) {
	w := testWriter()
	w.List()

	v := w.Value()
	v.String("hello, world")

	_, err := v.Build()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not root value")
}
