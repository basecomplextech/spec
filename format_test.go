package spec

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// list

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

		table0, size, err := _writeListTable(nil, ee0)
		if err != nil {
			t.Fatal(err)
		}

		table1 := readListTable(table0, size)
		ee1 := table1.elements()
		require.Equal(t, ee0, ee1)
	}
}

// offset

func TestListTable_offset__should_return_offset_by_index(t *testing.T) {
	elements := testListElements()
	data, size, err := _writeListTable(nil, elements)
	if err != nil {
		t.Fatal(err)
	}

	table := readListTable(data, size)

	for i, elem := range elements {
		off := table.offset(i)
		require.Equal(t, elem.offset, uint32(off))
	}
}

func TestListTable_offset__should_return_minus_one_when_out_of_range(t *testing.T) {
	elements := testListElements()
	data, size, err := _writeListTable(nil, elements)
	if err != nil {
		t.Fatal(err)
	}

	table := readListTable(data, size)

	off := table.offset(-1)
	assert.Equal(t, -1, off)

	n := table.count()
	off = table.offset(n)
	assert.Equal(t, -1, off)
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

// message

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
	data, size, err := _writeMessageTable(nil, fields)
	if err != nil {
		t.Fatal(err)
	}

	table := readMessageTable(data, size)

	n := table.count()
	assert.Equal(t, 10, n)
}

// read

func TestReadMessageTable__should_read_field_table(t *testing.T) {
	fields := testMessageFields()

	for i := 0; i <= len(fields); i++ {
		fields0 := fields[i:]

		data, size, err := _writeMessageTable(nil, fields0)
		if err != nil {
			t.Fatal(err)
		}

		table1 := readMessageTable(data, size)

		fields1 := table1.fields()
		require.Equal(t, fields0, fields1)
	}
}

// field

func TestMessageTable_field__should_return_field_by_index(t *testing.T) {
	fields := testMessageFields()
	data, size, err := _writeMessageTable(nil, fields)
	if err != nil {
		t.Fatal(err)
	}

	table := readMessageTable(data, size)

	for i, field := range fields {
		field1, ok := table.field(i)
		assert.True(t, ok)
		require.Equal(t, field, field1)
	}
}

func TestMessageTable_field__should_return_false_when_index_out_of_range(t *testing.T) {
	fields := testMessageFields()
	data, size, err := _writeMessageTable(nil, fields)
	if err != nil {
		t.Fatal(err)
	}

	table := readMessageTable(data, size)

	_, ok := table.field(-1)
	assert.False(t, ok)

	n := table.count()
	_, ok = table.field(n)
	assert.False(t, ok)
}

// offset

func TestMessageTable_offset__should_return_field_offset_by_tag(t *testing.T) {
	fields := testMessageFields()
	data, size, err := _writeMessageTable(nil, fields)
	if err != nil {
		t.Fatal(err)
	}

	table := readMessageTable(data, size)

	for _, field := range fields {
		off := table.offset(field.tag)
		require.Equal(t, field.offset, uint32(off))
	}
}

func TestMessageTable_offset__should_return_minus_one_when_field_not_found(t *testing.T) {
	fields := testMessageFields()
	data, size, err := _writeMessageTable(nil, fields)
	if err != nil {
		t.Fatal(err)
	}

	table := readMessageTable(data, size)

	off := table.offset(0)
	assert.Equal(t, -1, off)

	off = table.offset(math.MaxUint16)
	assert.Equal(t, -1, off)
}

// stack

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
