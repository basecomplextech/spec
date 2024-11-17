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

// Uint16

func TestDecodeUint16__should_decode_int16(t *testing.T) {
	b := buffer.New()
	encode.EncodeUint16(b, math.MaxUint16)
	p := b.Bytes()

	v, n, err := DecodeUint16(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, uint16(math.MaxUint16), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeUint16, typ)
	assert.Equal(t, size, len(p))
}

// Uint32

func TestDecodeUint32__should_decode_int32(t *testing.T) {
	b := buffer.New()
	encode.EncodeUint32(b, math.MaxUint32)
	p := b.Bytes()

	v, n, err := DecodeUint32(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, uint32(math.MaxUint32), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeUint32, typ)
	assert.Equal(t, size, len(p))
}

// Uint64

func TestDecodeUint64__should_decode_int64(t *testing.T) {
	b := buffer.New()
	encode.EncodeUint64(b, math.MaxUint64)
	p := b.Bytes()

	v, n, err := DecodeUint64(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, uint64(math.MaxUint64), v)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeUint64, typ)
	assert.Equal(t, size, len(p))
}

func TestDecodeUint64__should_decode_uint64_from_uint32(t *testing.T) {
	b := buffer.New()
	encode.EncodeUint32(b, math.MaxUint32)
	p := b.Bytes()

	v, n, err := DecodeUint64(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, uint64(math.MaxUint32), v)
}
