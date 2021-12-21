package spec

import "testing"

func BenchmarkFieldTable_field(b *testing.B) {
	fields := testMessageFieldsN(100)
	data, size, err := _writeMessageTable(nil, fields, false)
	if err != nil {
		b.Fatal(err)
	}

	table, err := _readMessageTable(data, size, false)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	last := len(fields) - 1
	for i := 0; i < b.N; i++ {
		f, ok := table.field(false, last)
		if !ok || f.tag == 0 || f.offset == 0 {
			b.Fatal()
		}
	}
}

func BenchmarkFieldTable_offset(b *testing.B) {
	fields := testMessageFieldsN(100)
	data, size, err := _writeMessageTable(nil, fields, false)
	if err != nil {
		b.Fatal(err)
	}

	table, err := _readMessageTable(data, size, false)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	last := len(fields) - 1
	tag := fields[last].tag

	for i := 0; i < b.N; i++ {
		start, _ := table.offset(false, tag)
		if start < 0 {
			b.Fatal()
		}
	}
}
