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

// DecodeBool

func TestDecodeBool__should_decode_bool_value(t *testing.T) {
	b := []byte{byte(format.TypeTrue)}
	v, n, err := DecodeBool(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, true, v)

	b = []byte{byte(format.TypeFalse)}
	v, n, err = DecodeBool(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, false, v)

	typ, size, err := DecodeTypeSize(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeFalse, typ)
	assert.Equal(t, size, len(b))
}

// DecodeByte

func TestDecodeByte__should_decode_byte(t *testing.T) {
	b := buffer.New()
	encode.EncodeByte(b, 1)
	p := b.Bytes()

	v, n, err := DecodeByte(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, byte(1), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeByte, typ)
	assert.Equal(t, size, len(p))
}
