package encoding

import (
	"math"
	"testing"

	"github.com/complex1tech/baselibrary/buffer"
	"github.com/complex1tech/baselibrary/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testEncodeListMeta(t tests.T, dataSize int, elements []ListElement) []byte {
	buf := buffer.New()
	buf.Grow(dataSize)

	_, err := EncodeListMeta(buf, dataSize, elements)
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// isBigList

func TestIsBigList__should_return_true_when_len_greater_than_uint8(t *testing.T) {
	smallTable := TestElementsN(math.MaxUint8)
	bigTable := TestElementsN(math.MaxUint8 + 1)

	assert.False(t, isBigList(smallTable))
	assert.True(t, isBigList(bigTable))
}

func TestIsBigList__should_return_true_when_offset_greater_than_uint16(t *testing.T) {
	smallTable := TestElementsSizeN(false, 1)
	bigTable := TestElementsSizeN(true, 1)

	assert.False(t, isBigList(smallTable))
	assert.True(t, isBigList(bigTable))
}

// len

func TestListTable_len_big__should_return_number_of_elements(t *testing.T) {
	big := true
	elements := TestElementsSize(big)

	buf := buffer.New()
	size, err := encodeListTable(buf, elements, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := decodeListTable(buf.Bytes(), uint32(size), big)
	if err != nil {
		t.Fatal(err)
	}

	n := table.len(big)
	assert.Equal(t, len(elements), n)
}

func TestListTable_len_smal__should_return_number_of_elements(t *testing.T) {
	small := false
	elements := TestElementsSize(small)

	buf := buffer.New()
	size, err := encodeListTable(buf, elements, small)
	if err != nil {
		t.Fatal(err)
	}
	table, err := decodeListTable(buf.Bytes(), uint32(size), small)
	if err != nil {
		t.Fatal(err)
	}

	n := table.len(small)
	assert.Equal(t, len(elements), n)
}

// offset: big

func TestListTable_offset_big__should_return_start_end_offset_by_index(t *testing.T) {
	big := true
	elements := TestElementsSize(big)

	buf := buffer.New()
	size, err := encodeListTable(buf, elements, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := decodeListTable(buf.Bytes(), uint32(size), big)
	if err != nil {
		t.Fatal(err)
	}

	for i, elem := range elements {
		prev := 0
		if i > 0 {
			_, prev = table.offset_big(i - 1)
		}

		start, end := table.offset_big(i)
		require.Equal(t, prev, start)
		require.Equal(t, int(elem.Offset), end)
	}
}

func TestListTable_offset_big__should_return_minus_one_when_out_of_range(t *testing.T) {
	big := true
	elements := TestElementsSize(big)

	buf := buffer.New()
	size, err := encodeListTable(buf, elements, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := decodeListTable(buf.Bytes(), uint32(size), big)
	if err != nil {
		t.Fatal(err)
	}

	start, end := table.offset_big(-1)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)

	n := table.len(big)
	start, end = table.offset_big(n)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)
}

// offset: small

func TestListTable_offset_small__should_return_start_end_offset_by_index(t *testing.T) {
	big := false
	elements := TestElementsSize(big)

	buf := buffer.New()
	size, err := encodeListTable(buf, elements, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := decodeListTable(buf.Bytes(), uint32(size), big)
	if err != nil {
		t.Fatal(err)
	}

	for i, elem := range elements {
		prev := 0
		if i > 0 {
			_, prev = table.offset_small(i - 1)
		}

		start, end := table.offset_small(i)
		require.Equal(t, prev, start)
		require.Equal(t, int(elem.Offset), end)
	}
}

func TestListTable_offset_small__should_return_minus_one_when_out_of_range(t *testing.T) {
	big := false
	elements := TestElementsSize(big)

	buf := buffer.New()
	size, err := encodeListTable(buf, elements, big)
	if err != nil {
		t.Fatal(err)
	}
	table, err := decodeListTable(buf.Bytes(), uint32(size), big)
	if err != nil {
		t.Fatal(err)
	}

	start, end := table.offset_small(-1)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)

	n := table.len(big)
	start, end = table.offset_small(n)
	assert.Equal(t, -1, start)
	assert.Equal(t, -1, end)
}
