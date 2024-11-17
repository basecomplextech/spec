// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package decode

import (
	"testing"

	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/basecomplextech/spec/internal/encode"
	"github.com/basecomplextech/spec/internal/format"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testEncodeMessageTable(t tests.T, dataSize int, fields []format.MessageField) []byte {
	buf := buffer.New()
	buf.Grow(dataSize)

	_, err := encode.EncodeMessageTable(buf, dataSize, fields)
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// Message

func TestDecodeMessageTable__should_decode_message_meta(t *testing.T) {
	fields := format.TestFields()
	dataSize := 100
	b := testEncodeMessageTable(t, dataSize, fields)

	meta, n, err := DecodeMessageTable(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(b), n)
	assert.Equal(t, uint32(dataSize), meta.DataSize())
	assert.Equal(t, len(fields), meta.Len())

	typ, size, err := DecodeTypeSize(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, format.TypeMessage, typ)
	assert.Equal(t, size, len(b))
}

func TestDecodeMessageTable__should_decode_message_table(t *testing.T) {
	fields := format.TestFields()

	for i := 0; i <= len(fields); i++ {
		buf := buffer.New()
		fields0 := fields[i:]

		_, err := encode.EncodeMessageTable(buf, 0, fields0)
		if err != nil {
			t.Fatal(err)
		}

		table1, _, err := DecodeMessageTable(buf.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		fields1 := table1.Fields()
		require.Equal(t, fields0, fields1)
	}
}

func TestDecodeMessageTable__should_return_error_when_invalid_type(t *testing.T) {
	fields := format.TestFields()
	dataSize := 100

	b := testEncodeMessageTable(t, dataSize, fields)
	b[len(b)-1] = byte(format.TypeList)

	_, _, err := DecodeMessageTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid type")
}

func TestDecodeMessageTable__should_return_error_when_invalid_table_size(t *testing.T) {
	b := []byte{}
	b = append(b, 0xff)
	b = append(b, byte(format.TypeMessage))

	_, _, err := DecodeMessageTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table size")
}

func TestDecodeMessageTable__should_return_error_when_invalid_data_size(t *testing.T) {
	big := false
	b := []byte{}
	b = append(b, 0xff)
	b = appendSize(b, big, 1000)
	b = append(b, byte(format.TypeMessage))

	_, _, err := DecodeMessageTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid data size")
}

func TestDecodeMessageTable__should_return_error_when_invalid_table(t *testing.T) {
	buf := buffer.New()
	_, err := encode.EncodeMessageTable(buf, 0, nil) // TODO: big(true)
	if err != nil {
		t.Fatal(err)
	}

	big := false
	b := buf.Bytes()
	b = appendSize(b, big, 0)    // data size
	b = appendSize(b, big, 1000) // table size
	b = append(b, byte(format.TypeMessage))

	_, _, err = DecodeMessageTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table")
}

func TestDecodeMessageTable__should_return_error_when_invalid_data(t *testing.T) {
	buf := buffer.New()

	_, err := encode.EncodeMessageTable(buf, 0, nil) // TODO: big(true)
	if err != nil {
		t.Fatal(err)
	}

	big := false
	b := buf.Bytes()
	b = appendSize(b, big, 1000)
	b = appendSize(b, big, 0)
	b = append(b, byte(format.TypeMessage))

	_, _, err = DecodeMessageTable(b)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid data")
}
