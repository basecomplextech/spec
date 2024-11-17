// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package decode

import (
	"testing"

	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/encode"
	"github.com/basecomplextech/spec/internal/format"
	"github.com/stretchr/testify/assert"
)

// Bin64/128/256

func TestDecodeBin64__should_decode_bin64(t *testing.T) {
	b := buffer.New()
	v := bin.Random64()
	encode.EncodeBin64(b, v)
	p := b.Bytes()

	v1, n, err := DecodeBin64(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, v, v1)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeBin64, typ)
	assert.Equal(t, size, len(p))
}

func TestDecodeBin128__should_decode_bin128(t *testing.T) {
	b := buffer.New()
	v := bin.Random128()
	encode.EncodeBin128(b, v)
	p := b.Bytes()

	v1, n, err := DecodeBin128(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, v, v1)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeBin128, typ)
	assert.Equal(t, size, len(p))
}

func TestDecodeBin256__should_decode_bin256(t *testing.T) {
	b := buffer.New()
	v := bin.Random256()
	encode.EncodeBin256(b, v)
	p := b.Bytes()

	v1, n, err := DecodeBin256(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, n, b.Len())
	assert.Equal(t, v, v1)

	typ, size, err := DecodeTypeSize(p)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeBin256, typ)
	assert.Equal(t, size, len(p))
}
