package protocol

import (
	"bytes"
	"encoding/json"
	"testing"
)

func BenchmarkFieldTable_find(b *testing.B) {
	fields := testMessageFieldsN(100)
	table := writeMessageTable(fields)
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
	fields := testMessageFieldsN(100)
	table := writeMessageTable(fields)
	b.ReportAllocs()
	b.ResetTimer()

	last := len(fields) - 1
	tag := fields[last].tag

	var f messageField
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

func BenchmarkWrite(b *testing.B) {
	msg := newTestMessage()

	buf := make([]byte, 0, 4096)
	w := NewWriterBuffer(buf)

	{
		if err := msg.Write(w); err != nil {
			b.Fatal(err)
		}
		data, err := w.End()
		if err != nil {
			b.Fatal(err)
		}

		b.SetBytes(int64(len(data)))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := msg.Write(w); err != nil {
			b.Fatal(err)
		}

		data, err := w.End()
		if err != nil {
			b.Fatal(err)
		}
		if len(data) < 100 {
			b.Fatal(len(data))
		}

		w.Reset()
		w.data.buffer = buf[:0]
	}
}

func BenchmarkJSON(b *testing.B) {
	msg := newTestMessage()

	buf := make([]byte, 0, 4096)
	buffer := bytes.NewBuffer(buf)

	{
		data, err := json.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
		b.SetBytes(int64(len(data)))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e := json.NewEncoder(buffer)
		if err := e.Encode(msg); err != nil {
			b.Fatal(err)
		}
		if buffer.Len() < 100 {
			b.Fatal(buffer.Len())
		}

		// b.Fatal(buffer.Len())
		// b.Fatal(buffer.String())
		buffer.Reset()
	}
}
