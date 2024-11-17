// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package writer

import (
	"math"
	"testing"

	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/spec/encoding"
	"github.com/stretchr/testify/assert"
)

func TestListWriter__should_write_list(t *testing.T) {
	w := testWriter()

	list := w.List()
	list.Bool(true)
	list.Byte(1)

	list.Int32(math.MaxInt32)
	list.Int64(math.MaxInt64)

	list.Uint32(math.MaxUint32)
	list.Uint64(math.MaxUint64)

	list.Float32(math.MaxFloat32)
	list.Float64(math.MaxFloat64)

	list.Bin64(bin.Random64())
	list.Bin128(bin.Random128())
	list.Bin256(bin.Random256())

	list.String("hello world")
	list.Bytes([]byte("hello world"))

	list1 := list.List()
	list1.String("sublist")
	if err := list1.End(); err != nil {
		t.Fatal(err)
	}

	msg1 := list.Message()
	msg1.Field(1).String("submessage")
	if err := msg1.End(); err != nil {
		t.Fatal(err)
	}

	b, err := list.Build()
	if err != nil {
		t.Fatal(err)
	}

	_, n, err := encoding.DecodeListTable(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(b), n)
}
