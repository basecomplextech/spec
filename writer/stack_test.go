package writer

import (
	"math/rand"
	"testing"
	"unsafe"

	"github.com/complex1tech/spec/encoding"
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

	// build buffer
	for _, elements := range matrix {
		offset := buffer.offset()
		offsets = append(offsets, offset)

		// push
		for _, elem := range elements {
			buffer.push(elem)
		}
	}

	// check buffer
	for i := len(offsets) - 1; i >= 0; i-- {
		offset := offsets[i]

		// pop table
		ff := buffer.pop(offset)
		elements := matrix[i]

		// check table
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

	// build buffer
	for _, fields := range matrix {
		offset := buffer.offset()
		offsets = append(offsets, offset)

		// copy
		ff := make([]encoding.MessageField, len(fields))
		copy(ff, fields)

		// shuffle
		rand.Shuffle(len(ff), func(i, j int) {
			ff[j], ff[i] = ff[i], ff[j]
		})

		// insert
		for _, f := range ff {
			buffer.insert(offset, f)
		}
	}

	// check buffer
	for i := len(offsets) - 1; i >= 0; i-- {
		offset := offsets[i]

		// pop table
		ff := buffer.pop(offset)
		fields := matrix[i]

		// check table
		require.Equal(t, fields, ff)
	}
}
