// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package decode

import (
	"testing"

	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/encode"
	"github.com/basecomplextech/spec/internal/format"
	"github.com/stretchr/testify/assert"
)

// Bytes

func TestDecodeBytes__should_decode_bytes(t *testing.T) {
	v := []byte("hello, world")

	b := buffer.New()
	_, err := encode.EncodeBytes(b, v)
	if err != nil {
		t.Fatal(err)
	}
	p := b.Bytes()

	v1, n, err := DecodeBytes(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, v, v1.Unwrap())

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeBytes, typ)
	assert.Equal(t, size, len(p))
}
