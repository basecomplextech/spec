// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"testing"

	"github.com/basecomplextech/baselibrary/buffer"
)

// messageTable: field_big

func BenchmarkMessageTable_field_big(b *testing.B) {
	buf := buffer.NewSize(4096)
	fields := TestFieldsN(100)

	size, err := encodeMessageTable(buf, fields, true /* big */)
	if err != nil {
		b.Fatal(err)
	}

	data := buf.Bytes()
	table, err := decodeMessageTable(data, uint32(size), true /* big */)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	last := len(fields) - 1
	for i := 0; i < b.N; i++ {
		f, ok := table.field_big(last)
		if !ok || f.Tag == 0 || f.Offset == 0 {
			b.Fatal()
		}
	}
}

// messageTable: field_small

func BenchmarkMessageTable_field_small(b *testing.B) {
	buf := buffer.NewSize(4096)
	fields := TestFieldsN(100)

	size, err := encodeMessageTable(buf, fields, false /* not big */)
	if err != nil {
		b.Fatal(err)
	}

	data := buf.Bytes()
	table, err := decodeMessageTable(data, uint32(size), false /* not big */)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	last := len(fields) - 1
	for i := 0; i < b.N; i++ {
		f, ok := table.field_small(last)
		if !ok || f.Tag == 0 || f.Offset == 0 {
			b.Fatal()
		}
	}

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec
	b.ReportMetric(ops/1000_000, "mops")
}

// messageTable: offset_big

func BenchmarkMessageTable_offset_big(b *testing.B) {
	buf := buffer.NewSize(4096)
	fields := TestFieldsN(100)

	size, err := encodeMessageTable(buf, fields, true /* big */)
	if err != nil {
		b.Fatal(err)
	}

	data := buf.Bytes()
	table, err := decodeMessageTable(data, uint32(size), true /* big */)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	last := len(fields) - 1
	tag := fields[last].Tag
	offset := int(fields[last].Offset)

	for i := 0; i < b.N; i++ {
		end := table.offset_big(tag)
		if end != offset {
			b.Fatal()
		}
	}

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec
	b.ReportMetric(ops/1000_000, "mops")
}

func BenchmarkMessageTable_offset_big_safe(b *testing.B) {
	buf := buffer.NewSize(4096)
	fields := TestFieldsN(100)

	size, err := encodeMessageTable(buf, fields, true)
	if err != nil {
		b.Fatal(err)
	}

	data := buf.Bytes()
	table, err := decodeMessageTable(data, uint32(size), true)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	last := len(fields) - 1
	tag := fields[last].Tag
	offset := int(fields[last].Offset)

	for i := 0; i < b.N; i++ {
		end := table.offset_big_safe(tag)
		if end != offset {
			b.Fatal()
		}
	}

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec
	b.ReportMetric(ops/1000_000, "mops")
}

// messageTable: offset_small

func BenchmarkMessageTable_offset_small(b *testing.B) {
	buf := buffer.NewSize(4096)
	fields := TestFieldsN(100)

	size, err := encodeMessageTable(buf, fields, false /* not big */)
	if err != nil {
		b.Fatal(err)
	}

	data := buf.Bytes()
	table, err := decodeMessageTable(data, uint32(size), false /* not big */)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	last := len(fields) - 1
	tag := fields[last].Tag
	offset := int(fields[last].Offset)

	for i := 0; i < b.N; i++ {
		end := table.offset_small(tag)
		if end != offset {
			b.Fatal()
		}
	}

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec
	b.ReportMetric(ops/1000_000, "mops")
}

func BenchmarkMessageTable_offset_small_safe(b *testing.B) {
	buf := buffer.NewSize(4096)
	fields := TestFieldsN(100)

	size, err := encodeMessageTable(buf, fields, false /* not big */)
	if err != nil {
		b.Fatal(err)
	}

	data := buf.Bytes()
	table, err := decodeMessageTable(data, uint32(size), false /* not big */)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	last := len(fields) - 1
	tag := fields[last].Tag
	offset := int(fields[last].Offset)

	for i := 0; i < b.N; i++ {
		end := table.offset_small_safe(tag)
		if end != offset {
			b.Fatal()
		}
	}

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec
	b.ReportMetric(ops/1000_000, "mops")
}
