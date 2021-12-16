package spec

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testMessageFields() []messageField {
	return testMessageFieldsN(10)
}

func testMessageFieldsN(n int) []messageField {
	result := make([]messageField, 0, n)
	for i := 0; i < n; i++ {
		field := messageField{
			tag:    uint16(i + 1),
			offset: uint32(i * 10),
		}
		result = append(result, field)
	}
	return result
}

// write/read

func TestMessageTable_write_read__should_write_and_read_field_table(t *testing.T) {
	fields := testMessageFields()

	for i := 0; i <= len(fields); i++ {
		fields0 := fields[i:]

		table0 := writeMessageTable(fields0)
		table1, err := readMessageTable(table0)
		if err != nil {
			t.Fatal(err)
		}

		fields1 := table1.fields()
		require.Equal(t, fields0, fields1)
	}
}

// get

func TestMessageTable_get__should_return_field_by_index(t *testing.T) {
	fields := testMessageFields()
	table := writeMessageTable(fields)

	for i, field := range fields {
		field1, ok := table.get(i)
		assert.True(t, ok)
		require.Equal(t, field, field1)
	}
}

func TestMessageTable_get__should_return_false_when_index_out_of_range(t *testing.T) {
	fields := testMessageFields()
	table := writeMessageTable(fields)

	_, ok := table.get(-1)
	assert.False(t, ok)

	n := table.count()
	_, ok = table.get(n)
	assert.False(t, ok)
}

// find

func TestMessageTable_find__should_find_field_by_tag(t *testing.T) {
	fields := testMessageFields()
	table := writeMessageTable(fields)

	for i, field := range fields {
		i1 := table.find(field.tag)
		require.Equal(t, i, i1)
	}
}

func TestMessageTable_find__should_return_minus_one_when_field_not_found(t *testing.T) {
	fields := testMessageFields()
	table := writeMessageTable(fields)

	i := table.find(0)
	assert.Equal(t, -1, i)

	i = table.find(math.MaxUint16)
	assert.Equal(t, -1, i)
}

// lookup

func TestMessageTable_lookup__should_find_field_by_tag(t *testing.T) {
	fields := testMessageFields()
	table := writeMessageTable(fields)

	for _, field := range fields {
		field1, ok := table.lookup(field.tag)
		require.Equal(t, field, field1)
		require.True(t, ok)
	}
}

func TestMessageTable_lookup__should_return_false_when_field_not_found(t *testing.T) {
	fields := testMessageFields()
	table := writeMessageTable(fields)

	_, ok := table.lookup(0)
	assert.False(t, ok)

	_, ok = table.lookup(math.MaxUint16)
	assert.False(t, ok)
}

// Count

func TestMessageTable_count__should_return_number_of_fields(t *testing.T) {
	fields := testMessageFieldsN(10)
	table := writeMessageTable(fields)

	n := table.count()
	assert.Equal(t, 10, n)
}

// Stack

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
