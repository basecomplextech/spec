package protocol

import "testing"

func BenchmarkFieldTable_find(b *testing.B) {
	fields := testFieldsN(100)
	table := writeFieldTable(fields)
	b.ReportAllocs()
	b.ResetTimer()

	last := len(fields) - 1
	tag := fields[last].tag
	for i := 0; i < b.N; i++ {
		j := table.find(tag)
		if j < 0 {
			b.Fatal()
		}
	}
}

func BenchmarkFieldTable_lookup(b *testing.B) {
	fields := testFieldsN(100)
	table := writeFieldTable(fields)
	b.ReportAllocs()
	b.ResetTimer()

	last := len(fields) - 1
	tag := fields[last].tag

	var f field
	var ok bool
	for i := 0; i < b.N; i++ {
		f, ok = table.lookup(tag)
		if !ok {
			b.Fatal()
		}
	}

	if f.tag == 0 {
		b.Fatal()
	}
}
