package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testElements() []element {
	return testElementsN(10)
}

func testElementsN(n int) []element {
	result := make([]element, 0, n)
	for i := 0; i < n; i++ {
		elem := element{
			offset: uint32(i * 10),
		}
		result = append(result, elem)
	}
	return result
}

// Write/Read

func TestElementTable_Write_Read__should_write_and_read_element_table(t *testing.T) {
	elements := testElements()

	for i := 0; i <= len(elements); i++ {
		ee0 := elements[i:]

		table0 := writeElementTable(ee0)
		table1, err := readElementTable(table0)
		if err != nil {
			t.Fatal(err)
		}

		ee1 := table1.elements()
		require.Equal(t, ee0, ee1)
	}
}

// Get

func TestElementTable_Get__should_return_element_by_index(t *testing.T) {
	elements := testElements()
	table := writeElementTable(elements)

	for i, elem := range elements {
		elem1 := table.get(i)
		require.Equal(t, elem, elem1)
	}
}

func TestElementTable_Get__should_panic_when_index_out_of_range(t *testing.T) {
	elements := testElements()
	table := writeElementTable(elements)

	assert.Panics(t, func() {
		table.get(-1)
	})

	assert.Panics(t, func() {
		n := table.count()
		table.get(n)
	})
}
