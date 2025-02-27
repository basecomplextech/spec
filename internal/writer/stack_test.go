// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package writer

import (
	"math/rand"
	"testing"
	"unsafe"

	"github.com/basecomplextech/spec/internal/format"
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
	matrix := [][]format.ListElement{
		format.TestElementsN(1),
		format.TestElementsN(10),
		format.TestElementsN(100),
		format.TestElementsN(10),
		format.TestElementsN(1),
		format.TestElementsN(0),
		format.TestElementsN(3),
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
	matrix := [][]format.MessageField{
		format.TestFieldsN(1),
		format.TestFieldsN(10),
		format.TestFieldsN(100),
		format.TestFieldsN(10),
		format.TestFieldsN(1),
		format.TestFieldsN(0),
		format.TestFieldsN(3),
	}

	buffer := messageStack{}
	offsets := []int{}

	// Build buffer
	for _, fields := range matrix {
		offset := buffer.offset()
		offsets = append(offsets, offset)

		// Copy
		ff := make([]format.MessageField, len(fields))
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
