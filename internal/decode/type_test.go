// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package decode

import (
	"testing"

	"github.com/basecomplextech/baselibrary/encoding/compactint"
	"github.com/basecomplextech/spec/internal/format"
	"github.com/stretchr/testify/assert"
)

func TestDecodeType__should_return_type(t *testing.T) {
	b := []byte{}
	b = append(b, byte(format.TypeString))

	v, n, err := DecodeType(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, n, len(b))
	assert.Equal(t, v, format.TypeString)
}

func TestDecodeType__should_return_undefined_when_empty(t *testing.T) {
	b := []byte{}

	v, n, err := DecodeType(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Zero(t, n)
	assert.Equal(t, v, format.TypeUndefined)
}

// util

// appendSize appends size as compactint, for tests.
func appendSize(b []byte, big bool, size uint32) []byte {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], size)
	off := compactint.MaxLen32 - n

	return append(b, p[off:]...)
}
