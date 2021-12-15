package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testListElements() []listElement {
	return testListElementsN(10)
}

func testListElementsN(n int) []listElement {
	result := make([]listElement, 0, n)
	for i := 0; i < n; i++ {
		elem := listElement{
			offset: uint32(i * 10),
		}
		result = append(result, elem)
	}
	return result
}

// write/read

func TestListTable_write_read__should_write_and_read_element_table(t *testing.T) {
	elements := testListElements()

	for i := 0; i <= len(elements); i++ {
		ee0 := elements[i:]

		table0 := writeListTable(ee0)
		table1, err := readListTable(table0)
		if err != nil {
			t.Fatal(err)
		}

		ee1 := table1.elements()
		require.Equal(t, ee0, ee1)
	}
}

// get

func TestListTable_get__should_return_element_by_index(t *testing.T) {
	elements := testListElements()
	table := writeListTable(elements)

	for i, elem := range elements {
		elem1 := table.get(i)
		require.Equal(t, elem, elem1)
	}
}

func TestListTable_get__should_panic_when_index_out_of_range(t *testing.T) {
	elements := testListElements()
	table := writeListTable(elements)

	assert.Panics(t, func() {
		table.get(-1)
	})

	assert.Panics(t, func() {
		n := table.count()
		table.get(n)
	})
}

// Stack

func TestListStack_Push__should_append_element_to_last_list(t *testing.T) {
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
