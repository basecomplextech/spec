// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package writer

import (
	"testing"

	"github.com/basecomplextech/spec/internal/decode"
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

	s, _, err := decode.DecodeString(b)
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
