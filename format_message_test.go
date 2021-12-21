package spec

import (
	"math"
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

// count

func TestMessageTable_count__should_return_number_of_fields(t *testing.T) {
	fields := testMessageFieldsN(10)
	data, size, err := _writeMessageTable(nil, fields, false)
	if err != nil {
		t.Fatal(err)
	}

	table, err := _readMessageTable(data, size, false)
	if err != nil {
		t.Fatal(err)
	}

	n := table.count(false)
	assert.Equal(t, 10, n)
}

// read

func TestReadMessageTable__should_read_field_table(t *testing.T) {
	fields := testMessageFields()

	for i := 0; i <= len(fields); i++ {
		fields0 := fields[i:]

		data, size, err := _writeMessageTable(nil, fields0, false)
		if err != nil {
			t.Fatal(err)
		}

		table1, err := _readMessageTable(data, size, false)
		if err != nil {
			t.Fatal(err)
		}

		fields1 := table1.fields(false)
		require.Equal(t, fields0, fields1)
	}
}

// field

func TestMessageTable_field__should_return_field_by_index(t *testing.T) {
	fields := testMessageFields()
	data, size, err := _writeMessageTable(nil, fields, false)
	if err != nil {
		t.Fatal(err)
	}

	table, err := _readMessageTable(data, size, false)
	if err != nil {
		t.Fatal(err)
	}

	for i, field := range fields {
		field1, ok := table.field(false, i)
		assert.True(t, ok)
		require.Equal(t, field, field1)
	}
}

func TestMessageTable_field__should_return_false_when_index_out_of_range(t *testing.T) {
	fields := testMessageFields()
	data, size, err := _writeMessageTable(nil, fields, false)
	if err != nil {
		t.Fatal(err)
	}

	table, err := _readMessageTable(data, size, false)
	if err != nil {
		t.Fatal(err)
	}

	_, ok := table.field(false, -1)
	assert.False(t, ok)

	n := table.count(false)
	_, ok = table.field(false, n)
	assert.False(t, ok)
}

// offset

func TestMessageTable_offset__should_return_start_end_offset_by_tag(t *testing.T) {
	fields := testMessageFields()
	data, size, err := _writeMessageTable(nil, fields, false)
	if err != nil {
		t.Fatal(err)
	}

	table, err := _readMessageTable(data, size, false)
	if err != nil {
		t.Fatal(err)
	}

	for i, field := range fields {
		prev := 0
		if i > 0 {
			_, prev = table.offset(false, field.tag-1)
		}

		start, end := table.offset(false, field.tag)
		require.Equal(t, prev, start)
		require.Equal(t, int(field.offset), end)
	}
}

func TestMessageTable_offset__should_return_minus_one_when_field_not_found(t *testing.T) {
	fields := testMessageFields()
	data, size, err := _writeMessageTable(nil, fields, false)
	if err != nil {
		t.Fatal(err)
	}

	table, err := _readMessageTable(data, size, false)
	if err != nil {
		t.Fatal(err)
	}

	start, end := table.offset(false, 0)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)

	start, end = table.offset(false, math.MaxUint16)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)
}
