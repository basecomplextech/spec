package spec

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testMessageFields() []messageField {
	return testMessageFieldsSizeN(false, 10)
}

func testMessageFieldsN(n int) []messageField {
	return testMessageFieldsSizeN(false, n)
}

func testMessageFieldsSize(big bool) []messageField {
	return testMessageFieldsSizeN(big, 10)
}

func testMessageFieldsSizeN(big bool, n int) []messageField {
	tagStart := uint16(0)
	offStart := uint32(0)
	if big {
		tagStart = math.MaxUint8 + 1
		offStart = math.MaxUint16 + 1
	}

	result := make([]messageField, 0, n)
	for i := 0; i < n; i++ {
		field := messageField{
			tag:    tagStart + uint16(i+1),
			offset: offStart + uint32(i*10),
		}
		result = append(result, field)
	}
	return result
}

// isBigMessage

func TestIsBigMessage__should_return_true_when_tag_greater_than_uint8(t *testing.T) {
	small := testMessageFieldsSize(false)
	big := testMessageFieldsSize(true)

	// clear offsets to check tags
	for i, f := range small {
		f.offset = 0
		small[i] = f
	}
	for i, f := range big {
		f.offset = 0
		big[i] = f
	}

	assert.False(t, isBigMessage(small))
	assert.True(t, isBigMessage(big))
}

func TestIsBigMessage__should_return_true_when_offset_greater_than_uint16(t *testing.T) {
	small := testMessageFieldsSize(false)
	big := testMessageFieldsSize(true)

	// clear tags to check offsets
	for i, f := range small {
		f.tag = 0
		small[i] = f
	}
	for i, f := range big {
		f.tag = 0
		big[i] = f
	}

	assert.False(t, isBigMessage(small))
	assert.True(t, isBigMessage(big))
}

// count

func TestMessageTable_count_big__should_return_number_of_fields(t *testing.T) {
	big := true
	fields := testMessageFieldsSizeN(big, 10)

	data, size, err := _writeMessageTable(nil, fields, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readMessageTable(data, size, big)
	if err != nil {
		t.Fatal(err)
	}

	n := table.count(big)
	assert.Equal(t, 10, n)
}

func TestMessageTable_count_small__should_return_number_of_fields(t *testing.T) {
	small := false
	fields := testMessageFieldsSizeN(small, 10)

	data, size, err := _writeMessageTable(nil, fields, small)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readMessageTable(data, size, small)
	if err != nil {
		t.Fatal(err)
	}

	n := table.count(small)
	assert.Equal(t, 10, n)
}

// offset: big

func TestMessageTable_offset_big__should_return_start_end_offset_by_tag(t *testing.T) {
	big := true
	fields := testMessageFieldsSize(big)

	data, size, err := _writeMessageTable(nil, fields, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readMessageTable(data, size, big)
	if err != nil {
		t.Fatal(err)
	}

	for i, field := range fields {
		prev := 0
		if i > 0 {
			_, prev = table.offset(big, field.tag-1)
		}

		start, end := table.offset(big, field.tag)
		require.Equal(t, prev, start)
		require.Equal(t, int(field.offset), end)
	}
}

func TestMessageTable_offset_big__should_return_minus_one_when_field_not_found(t *testing.T) {
	big := true
	fields := testMessageFieldsSize(big)

	data, size, err := _writeMessageTable(nil, fields, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readMessageTable(data, size, big)
	if err != nil {
		t.Fatal(err)
	}

	start, end := table.offset(big, 0)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)

	start, end = table.offset(big, math.MaxUint16)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)
}

// offset: small

func TestMessageTable_offset_small__should_return_start_end_offset_by_tag(t *testing.T) {
	small := false
	fields := testMessageFieldsSize(small)

	data, size, err := _writeMessageTable(nil, fields, small)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readMessageTable(data, size, small)
	if err != nil {
		t.Fatal(err)
	}

	for i, field := range fields {
		prev := 0
		if i > 0 {
			_, prev = table.offset(small, field.tag-1)
		}

		start, end := table.offset(small, field.tag)
		require.Equal(t, prev, start)
		require.Equal(t, int(field.offset), end)
	}
}

func TestMessageTable_offset_small__should_return_minus_one_when_field_not_found(t *testing.T) {
	small := false
	fields := testMessageFieldsSize(small)

	data, size, err := _writeMessageTable(nil, fields, small)
	if err != nil {
		t.Fatal(err)
	}

	table, err := _readMessageTable(data, size, small)
	if err != nil {
		t.Fatal(err)
	}

	start, end := table.offset(small, 0)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)

	start, end = table.offset(small, math.MaxUint16)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)
}

// offsetByIndex: big

func TestMessageTable_offsetByIndex_big__should_return_start_end_offset_by_index(t *testing.T) {
	big := true
	fields := testMessageFieldsSize(big)

	data, size, err := _writeMessageTable(nil, fields, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readMessageTable(data, size, big)
	if err != nil {
		t.Fatal(err)
	}

	for i, field := range fields {
		prev := 0
		if i > 0 {
			_, prev = table.offsetByIndex(big, i-1)
		}

		start, end := table.offsetByIndex(big, i)
		require.Equal(t, prev, start)
		require.Equal(t, int(field.offset), end)
	}
}

func TestMessageTable_offsetByIndex_big__should_return_minus_one_when_field_not_found(t *testing.T) {
	big := true
	fields := testMessageFieldsSize(big)

	data, size, err := _writeMessageTable(nil, fields, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readMessageTable(data, size, big)
	if err != nil {
		t.Fatal(err)
	}

	start, end := table.offsetByIndex(big, -1)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)

	start, end = table.offsetByIndex(big, math.MaxUint16)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)
}

// offsetByIndex: small

func TestMessageTable_offsetByIndex_small__should_return_start_end_offset_by_index(t *testing.T) {
	small := false
	fields := testMessageFieldsSize(small)

	data, size, err := _writeMessageTable(nil, fields, small)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readMessageTable(data, size, small)
	if err != nil {
		t.Fatal(err)
	}

	for i, field := range fields {
		prev := 0
		if i > 0 {
			_, prev = table.offsetByIndex(small, i-1)
		}

		start, end := table.offsetByIndex(small, i)
		require.Equal(t, prev, start)
		require.Equal(t, int(field.offset), end)
	}
}

func TestMessageTable_offsetByIndex_small__should_return_minus_one_when_field_not_found(t *testing.T) {
	small := false
	fields := testMessageFieldsSize(small)

	data, size, err := _writeMessageTable(nil, fields, small)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readMessageTable(data, size, small)
	if err != nil {
		t.Fatal(err)
	}

	start, end := table.offsetByIndex(small, -1)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)

	start, end = table.offsetByIndex(small, math.MaxUint16)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)
}

// field: big

func TestMessageTable_field_big__should_return_field_by_index(t *testing.T) {
	big := true
	fields := testMessageFieldsSize(big)

	data, size, err := _writeMessageTable(nil, fields, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readMessageTable(data, size, big)
	if err != nil {
		t.Fatal(err)
	}

	for i, field := range fields {
		field1, ok := table.field(big, i)
		assert.True(t, ok)
		require.Equal(t, field, field1)
	}
}

func TestMessageTable_field_big__should_return_false_when_index_out_of_range(t *testing.T) {
	big := true
	fields := testMessageFieldsSize(big)

	data, size, err := _writeMessageTable(nil, fields, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readMessageTable(data, size, big)
	if err != nil {
		t.Fatal(err)
	}

	_, ok := table.field(big, -1)
	assert.False(t, ok)

	n := table.count(big)
	_, ok = table.field(big, n)
	assert.False(t, ok)
}

// field: small

func TestMessageTable_field_small__should_return_field_by_index(t *testing.T) {
	small := false
	fields := testMessageFieldsSize(small)

	data, size, err := _writeMessageTable(nil, fields, small)
	if err != nil {
		t.Fatal(err)
	}

	table, err := _readMessageTable(data, size, small)
	if err != nil {
		t.Fatal(err)
	}

	for i, field := range fields {
		field1, ok := table.field(small, i)
		assert.True(t, ok)
		require.Equal(t, field, field1)
	}
}

func TestMessageTable_field_small__should_return_false_when_index_out_of_range(t *testing.T) {
	small := false
	fields := testMessageFieldsSize(small)

	data, size, err := _writeMessageTable(nil, fields, small)
	if err != nil {
		t.Fatal(err)
	}

	table, err := _readMessageTable(data, size, small)
	if err != nil {
		t.Fatal(err)
	}

	_, ok := table.field(small, -1)
	assert.False(t, ok)

	n := table.count(small)
	_, ok = table.field(small, n)
	assert.False(t, ok)
}
