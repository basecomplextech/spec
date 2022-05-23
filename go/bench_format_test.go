package spec

import (
	"testing"

	"github.com/epochtimeout/basekit/system/buffer"
)

func BenchmarkFieldTable_field(b *testing.B) {
	buf := buffer.NewSize(4096)
	fields := testMessageFieldsN(100)

	size, err := encodeMessageTable(buf, fields, false)
	if err != nil {
		b.Fatal(err)
	}

	data := buf.Bytes()
	table, err := decodeMessageTable(data, uint32(size), false)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	last := len(fields) - 1
	for i := 0; i < b.N; i++ {
		f, ok := table.field_small(last)
		if !ok || f.tag == 0 || f.offset == 0 {
			b.Fatal()
		}
	}
}

func BenchmarkFieldTable_offset(b *testing.B) {
	buf := buffer.NewSize(4096)
	fields := testMessageFieldsN(100)

	size, err := encodeMessageTable(buf, fields, false)
	if err != nil {
		b.Fatal(err)
	}

	data := buf.Bytes()
	table, err := decodeMessageTable(data, uint32(size), false)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	last := len(fields) - 1
	tag := fields[last].tag

	for i := 0; i < b.N; i++ {
		end := table.offset_small(tag)
		if end < 0 {
			b.Fatal()
		}
	}
}
