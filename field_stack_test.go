package protocol

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFieldStack_Insert__should_insert_field_into_table_ordered_by_tags(t *testing.T) {
	matrix := [][]field{
		testFieldsN(1),
		testFieldsN(10),
		testFieldsN(100),
		testFieldsN(10),
		testFieldsN(1),
		testFieldsN(0),
		testFieldsN(3),
	}

	stack := fieldStack{}
	offsets := []int{}

	// build stack
	for _, fields := range matrix {
		offset := stack.offset()
		offsets = append(offsets, offset)

		// copy
		ff := make([]field, len(fields))
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
		ff := stack.popTable(offset)
		fields := matrix[i]

		// check table
		require.Equal(t, fields, ff)
	}
}