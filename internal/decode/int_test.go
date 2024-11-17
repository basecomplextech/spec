// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package decode

import (
	"math"
	"testing"

	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/encode"
	"github.com/basecomplextech/spec/internal/format"
	"github.com/stretchr/testify/assert"
)

// Int16

func TestDecodeInt16__should_decode_int16(t *testing.T) {
	b := buffer.New()
	encode.EncodeInt16(b, math.MaxInt16)
	p := b.Bytes()

	v, n, err := DecodeInt16(p)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, b.Len())
	assert.Equal(t, int16(math.MaxInt16), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeInt16, typ)
	assert.Equal(t, size, len(p))
}

// Int32

func TestDecodeInt32__should_decode_int32(t *testing.T) {
	b := buffer.New()
	encode.EncodeInt32(b, math.MaxInt32)
	p := b.Bytes()

	v, n, err := DecodeInt32(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, int32(math.MaxInt32), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeInt32, typ)
	assert.Equal(t, size, len(p))
}

// Int64

func TestDecodeInt64__should_decode_int64(t *testing.T) {
	b := buffer.New()
	encode.EncodeInt64(b, math.MaxInt64)
	p := b.Bytes()

	v, n, err := DecodeInt64(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, int64(math.MaxInt64), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeInt64, typ)
	assert.Equal(t, size, len(p))
}

func TestDecodeInt64__should_decode_int64_from_int32(t *testing.T) {
	b := buffer.New()
	encode.EncodeInt32(b, math.MaxInt32)
	p := b.Bytes()

	v, n, err := DecodeInt64(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, int64(math.MaxInt32), v)
}
