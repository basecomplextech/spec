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

// Float32

func TestDecodeFloat32__should_decode_float32(t *testing.T) {
	b := buffer.New()
	encode.EncodeFloat32(b, math.MaxFloat32)
	p := b.Bytes()

	v, n, err := DecodeFloat32(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, float32(math.MaxFloat32), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeFloat32, typ)
	assert.Equal(t, size, len(p))
}

func TestDecodeFloat32__should_decode_float32_from_float64(t *testing.T) {
	b := buffer.New()
	encode.EncodeFloat64(b, math.MaxFloat32)
	p := b.Bytes()

	v, n, err := DecodeFloat32(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, float32(math.MaxFloat32), v)
}

// Float64

func TestDecodeFloat64__should_decode_float64(t *testing.T) {
	b := buffer.New()
	encode.EncodeFloat64(b, math.MaxFloat64)
	p := b.Bytes()

	v, n, err := DecodeFloat64(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, float64(math.MaxFloat64), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeFloat64, typ)
	assert.Equal(t, size, len(p))
}

func TestDecodeFloat64__should_decode_float64_from_float32(t *testing.T) {
	b := buffer.New()
	encode.EncodeFloat32(b, math.MaxFloat32)
	p := b.Bytes()

	v, n, err := DecodeFloat64(p)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, b.Len())
	assert.Equal(t, float64(math.MaxFloat32), v)
}
