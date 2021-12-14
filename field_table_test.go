package protocol

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testFields() []field {
	return testFieldsN(10)
}

func testFieldsN(n int) []field {
	result := make([]field, 0, n)
	for i := 0; i < n; i++ {
		field := field{
			tag:    uint16(i + 1),
			offset: uint32(i * 10),
		}
		result = append(result, field)
	}
	return result
}

// Write/Read

func TestFieldTable_Write_Read__should_write_and_read_field_table(t *testing.T) {
	fields := testFields()

	for i := 0; i <= len(fields); i++ {
		fields0 := fields[i:]

		table0 := writeFieldTable(fields0)
		table1, err := readFieldTable(table0)
		if err != nil {
			t.Fatal(err)
		}

		fields1 := table1.fields()
		require.Equal(t, fields0, fields1)
	}
}

// Get

func TestFieldTable_Get__should_return_field_by_index(t *testing.T) {
	fields := testFields()
	table := writeFieldTable(fields)

	for i, field := range fields {
		field1 := table.get(i)
		require.Equal(t, field, field1)
	}
}

func TestFieldTable_Get__should_panic_when_index_out_of_range(t *testing.T) {
	fields := testFields()
	table := writeFieldTable(fields)

	assert.Panics(t, func() {
		table.get(-1)
	})

	assert.Panics(t, func() {
		n := table.count()
		table.get(n)
	})
}

// Find

func TestFieldTable_Find__should_find_field_by_tag(t *testing.T) {
	fields := testFields()
	table := writeFieldTable(fields)

	for i, field := range fields {
		i1 := table.find(field.tag)
		require.Equal(t, i, i1)
	}
}

func TestFieldTable_Find__should_return_minus_one_when_field_not_found(t *testing.T) {
	fields := testFields()
	table := writeFieldTable(fields)

	i := table.find(0)
	assert.Equal(t, -1, i)

	i = table.find(math.MaxUint16)
	assert.Equal(t, -1, i)
}

// Lookup

func TestFieldTable_Lookup__should_find_field_by_tag(t *testing.T) {
	fields := testFields()
	table := writeFieldTable(fields)

	for _, field := range fields {
		field1, ok := table.lookup(field.tag)
		require.Equal(t, field, field1)
		require.True(t, ok)
	}
}

func TestFieldTable_Lookup__should_return_false_when_field_not_found(t *testing.T) {
	fields := testFields()
	table := writeFieldTable(fields)

	_, ok := table.lookup(0)
	assert.False(t, ok)

	_, ok = table.lookup(math.MaxUint16)
	assert.False(t, ok)
}

// Count

func TestFieldTable_Count__should_return_number_of_fields(t *testing.T) {
	fields := testFieldsN(10)
	table := writeFieldTable(fields)

	n := table.count()
	assert.Equal(t, 10, n)
}
