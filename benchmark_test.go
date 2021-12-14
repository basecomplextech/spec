package protocol

import "testing"

func BenchmarkFieldTable_Lookup(b *testing.B) {
	fields := testFieldsN(1000)
	table := writeFieldTable(fields)
	b.ReportAllocs()
	b.ResetTimer()

	tag := fields[0].tag
	for i := 0; i < b.N; i++ {
		_, ok := table.lookup(tag)
		if !ok {
			b.Fatal()
		}
	}
}
