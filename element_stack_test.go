package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestElementStack_Push__should_append_element_to_last_list(t *testing.T) {
	matrix := [][]element{
		testElementsN(1),
		testElementsN(10),
		testElementsN(100),
		testElementsN(10),
		testElementsN(1),
		testElementsN(0),
		testElementsN(3),
	}

	stack := elementStack{}
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

		// pop list
		ff := stack.popList(offset)
		elements := matrix[i]

		// check list
		require.Equal(t, elements, ff)
	}
}
