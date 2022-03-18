package spec

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testListElements() []listElement {
	return testListElementsN(10)
}

func testListElementsN(n int) []listElement {
	return testListElementsSizeN(false, n)
}

func testListElementsSize(big bool) []listElement {
	return testListElementsSizeN(big, 10)
}

func testListElementsSizeN(big bool, n int) []listElement {
	start := uint32(0)
	if big {
		start = math.MaxUint16 + 1
	}

	result := make([]listElement, 0, n)
	for i := 0; i < n; i++ {
		elem := listElement{
			offset: start + uint32(i*10),
		}
		result = append(result, elem)
	}
	return result
}

// isBigList

func TestIsBigList__should_return_true_when_count_greater_than_uint8(t *testing.T) {
	smallTable := testListElementsN(math.MaxUint8)
	bigTable := testListElementsN(math.MaxUint8 + 1)

	assert.False(t, isBigList(smallTable))
	assert.True(t, isBigList(bigTable))
}

func TestIsBigList__should_return_true_when_offset_greater_than_uint16(t *testing.T) {
	smallTable := testListElementsSizeN(false, 1)
	bigTable := testListElementsSizeN(true, 1)

	assert.False(t, isBigList(smallTable))
	assert.True(t, isBigList(bigTable))
}

// count

func TestListTable_count_big__should_return_number_of_elements(t *testing.T) {
	big := true
	elements := testListElementsSize(big)

	data, size, err := encodeListTable(nil, elements, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readListTable(data, size, big)
	if err != nil {
		t.Fatal(err)
	}

	n := table.count(big)
	assert.Equal(t, len(elements), n)
}

func TestListTable_count_smal__should_return_number_of_elements(t *testing.T) {
	small := false
	elements := testListElementsSize(small)

	data, size, err := encodeListTable(nil, elements, small)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readListTable(data, size, small)
	if err != nil {
		t.Fatal(err)
	}

	n := table.count(small)
	assert.Equal(t, len(elements), n)
}

// offset: big

func TestListTable_offset_big__should_return_start_end_offset_by_index(t *testing.T) {
	big := true
	elements := testListElementsSize(big)

	data, size, err := encodeListTable(nil, elements, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readListTable(data, size, big)
	if err != nil {
		t.Fatal(err)
	}

	for i, elem := range elements {
		prev := 0
		if i > 0 {
			_, prev = table.offset(big, i-1)
		}

		start, end := table.offset(big, i)
		require.Equal(t, prev, start)
		require.Equal(t, int(elem.offset), end)
	}
}

func TestListTable_offset_big__should_return_minus_one_when_out_of_range(t *testing.T) {
	big := true
	elements := testListElementsSize(big)

	data, size, err := encodeListTable(nil, elements, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readListTable(data, size, big)
	if err != nil {
		t.Fatal(err)
	}

	start, end := table.offset(big, -1)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)

	n := table.count(big)
	start, end = table.offset(big, n)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)
}

// offset: small

func TestListTable_offset_small__should_return_start_end_offset_by_index(t *testing.T) {
	small := false
	elements := testListElementsSize(small)

	data, size, err := encodeListTable(nil, elements, small)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readListTable(data, size, small)
	if err != nil {
		t.Fatal(err)
	}

	for i, elem := range elements {
		prev := 0
		if i > 0 {
			_, prev = table.offset(small, i-1)
		}

		start, end := table.offset(small, i)
		require.Equal(t, prev, start)
		require.Equal(t, int(elem.offset), end)
	}
}

func TestListTable_offset_small__should_return_minus_one_when_out_of_range(t *testing.T) {
	small := false
	elements := testListElementsSize(small)

	data, size, err := encodeListTable(nil, elements, small)
	if err != nil {
		t.Fatal(err)
	}
	table, err := _readListTable(data, size, small)
	if err != nil {
		t.Fatal(err)
	}

	start, end := table.offset(small, -1)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)

	n := table.count(small)
	start, end = table.offset(small, n)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)
}
