package spec

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

// read

func TestReadListTable__should_read_list_table(t *testing.T) {
	elements := testListElements()

	for i := 0; i <= len(elements); i++ {
		ee0 := elements[i:]

		table0, size, err := _writeListTable(nil, ee0, false)
		if err != nil {
			t.Fatal(err)
		}

		table1, err := _readListTable(table0, size)
		if err != nil {
			t.Fatal(err)
		}

		ee1 := table1.elements(false)
		require.Equal(t, ee0, ee1)
	}
}

// offset

func TestListTable_offset__should_return_start_end_offset_by_index(t *testing.T) {
	elements := testListElements()
	data, size, err := _writeListTable(nil, elements, false)
	if err != nil {
		t.Fatal(err)
	}

	table, err := _readListTable(data, size)
	if err != nil {
		t.Fatal(err)
	}

	for i, elem := range elements {
		prev := 0
		if i > 0 {
			_, prev = table.offset(false, i-1)
		}

		start, end := table.offset(false, i)
		require.Equal(t, prev, start)
		require.Equal(t, int(elem.offset), end)
	}
}

func TestListTable_offset__should_return_minus_one_when_out_of_range(t *testing.T) {
	elements := testListElements()
	data, size, err := _writeListTable(nil, elements, false)
	if err != nil {
		t.Fatal(err)
	}

	table, err := _readListTable(data, size)
	if err != nil {
		t.Fatal(err)
	}

	start, end := table.offset(false, -1)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)

	n := table.count(false)
	start, end = table.offset(false, n)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)
}

// stack

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
