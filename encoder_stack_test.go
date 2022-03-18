package spec

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

// list list

func TestListStack_push__should_append_element_to_last_list(t *testing.T) {
	matrix := [][]listElement{
		testListElementsN(1),
		testListElementsN(10),
		testListElementsN(100),
		testListElementsN(10),
		testListElementsN(1),
		testListElementsN(0),
		testListElementsN(3),
	}

	stack := listStack{}
	offsets := []int{}

	// build stack
	for _, elements := range matrix {
		offset := stack.offset()
		offsets = append(offsets, offset)

		// push
		for _, elem := range elements {
			stack.push(elem)
		}
	}

	// check stack
	for i := len(offsets) - 1; i >= 0; i-- {
		offset := offsets[i]

		// pop table
		ff := stack.pop(offset)
		elements := matrix[i]

		// check table
		require.Equal(t, elements, ff)
	}
}

// message stack

func TestFieldStack_Insert__should_insert_field_into_table_ordered_by_tags(t *testing.T) {
	matrix := [][]messageField{
		testMessageFieldsN(1),
		testMessageFieldsN(10),
		testMessageFieldsN(100),
		testMessageFieldsN(10),
		testMessageFieldsN(1),
		testMessageFieldsN(0),
		testMessageFieldsN(3),
	}

	stack := messageStack{}
	offsets := []int{}

	// build stack
	for _, fields := range matrix {
		offset := stack.offset()
		offsets = append(offsets, offset)

		// copy
		ff := make([]messageField, len(fields))
		copy(ff, fields)

		// shuffle
		rand.Shuffle(len(ff), func(i, j int) {
			ff[j], ff[i] = ff[i], ff[j]
		})

		// insert
		for _, f := range ff {
			stack.insert(offset, f)
		}
	}

	// check stack
	for i := len(offsets) - 1; i >= 0; i-- {
		offset := offsets[i]

		// pop table
		ff := stack.pop(offset)
		fields := matrix[i]

		// check table
		require.Equal(t, fields, ff)
	}
}
