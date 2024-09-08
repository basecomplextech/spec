// Copyright 2021 Ivan Korobkov. All rights reserved.

package writer

import (
	"math/rand"
	"testing"
	"unsafe"

	"github.com/basecomplextech/spec/encoding"
	"github.com/stretchr/testify/require"
)

func TestWriterState_Size__should_be_lte_1kb(t *testing.T) {
	s := unsafe.Sizeof(writerState{})
	if s > 1024 {
		t.Fatal(s)
	}
}

// list buffer

func TestListBuffer_push__should_append_element_to_last_list(t *testing.T) {
	matrix := [][]encoding.ListElement{
		encoding.TestElementsN(1),
		encoding.TestElementsN(10),
		encoding.TestElementsN(100),
		encoding.TestElementsN(10),
		encoding.TestElementsN(1),
		encoding.TestElementsN(0),
		encoding.TestElementsN(3),
	}

	buffer := listStack{}
	offsets := []int{}

	// Build buffer
	for _, elements := range matrix {
		offset := buffer.offset()
		offsets = append(offsets, offset)

		// Push
		for _, elem := range elements {
			buffer.push(elem)
		}
	}

	// Check buffer
	for i := len(offsets) - 1; i >= 0; i-- {
		offset := offsets[i]

		// Pop table
		ff := buffer.pop(offset)
		elements := matrix[i]

		// Check table
		require.Equal(t, elements, ff)
	}
}

// message buffer

func TestMessageBuffer_Insert__should_insert_field_into_table_ordered_by_tags(t *testing.T) {
	matrix := [][]encoding.MessageField{
		encoding.TestFieldsN(1),
		encoding.TestFieldsN(10),
		encoding.TestFieldsN(100),
		encoding.TestFieldsN(10),
		encoding.TestFieldsN(1),
		encoding.TestFieldsN(0),
		encoding.TestFieldsN(3),
	}

	buffer := messageStack{}
	offsets := []int{}

	// Build buffer
	for _, fields := range matrix {
		offset := buffer.offset()
		offsets = append(offsets, offset)

		// Copy
		ff := make([]encoding.MessageField, len(fields))
		copy(ff, fields)

		// Shuffle
		rand.Shuffle(len(ff), func(i, j int) {
			ff[j], ff[i] = ff[i], ff[j]
		})

		// Insert
		for _, f := range ff {
			buffer.insert(offset, f)
		}
	}

	// Check buffer
	for i := len(offsets) - 1; i >= 0; i-- {
		offset := offsets[i]

		// Pop table
		ff := buffer.pop(offset)
		fields := matrix[i]

		// Check table
		require.Equal(t, fields, ff)
	}
}
