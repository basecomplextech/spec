package spec

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"testing"
	"time"
)

func Benchmark_Read(b *testing.B) {
	msg := newTestMessage()

	buf := make([]byte, 0, 4096)
	w := NewWriterBuffer(buf)
	if err := msg.Write(w); err != nil {
		b.Fatal(err)
	}
	data, err := w.End()
	if err != nil {
		b.Fatal(err)
	}
	reader := readTestMessageData(data)

	size := len(data)
	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		v := walkMessageData(reader)
		if v == 0 {
			b.Fatal(v)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(compressedSize(data)), "size-zlib")
}

// Write

func Benchmark_Write(b *testing.B) {
	msg := newTestMessage()

	buf := make([]byte, 0, 4096)
	size := int64(0)
	w := NewWriterBuffer(buf)
	{
		if err := msg.Write(w); err != nil {
			b.Fatal(err)
		}
		data, err := w.End()
		if err != nil {
			b.Fatal(err)
		}

		size = int64(len(data))
		b.SetBytes(size)
	}

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()

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

		// b.Fatal(len(data))
		w.Reset()
		w.buf = buf[:0]
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(size), "size")
}

// JSON

func BenchmarkJSON_Read(b *testing.B) {
	msg := newTestMessage()

	data, err := json.Marshal(msg)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()

	msg1 := &TestMessage{}
	for i := 0; i < b.N; i++ {
		if err := json.Unmarshal(data, msg1); err != nil {
			b.Fatal(err)
		}

		v := walkMessage(msg1)
		if v == 0 {
			b.Fatal(v)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(len(data)), "size")
	b.ReportMetric(float64(compressedSize(data)), "size-zlib")
}

func BenchmarkJSON_Write(b *testing.B) {
	msg := newTestMessage()

	buf := make([]byte, 0, 4096)
	buffer := bytes.NewBuffer(buf)
	size := int64(0)

	{
		data, err := json.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}

		size = int64(len(data))
		b.SetBytes(size)
	}

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()

	for i := 0; i < b.N; i++ {
		e := json.NewEncoder(buffer)
		if err := e.Encode(msg); err != nil {
			b.Fatal(err)
		}
		if buffer.Len() < 100 {
			b.Fatal(buffer.Len())
		}

		// b.Fatal(buffer.Len(), buffer.String())
		buffer.Reset()
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(size), "size")
}

// private

func walkMessage(m *TestMessage) int {
	var v int

	v += int(m.Int8)
	v += int(m.Int16)
	v += int(m.Int32)
	v += int(m.Int64)

	v += int(m.UInt8)
	v += int(m.UInt16)
	v += int(m.UInt32)
	v += int(m.UInt64)

	v += int(m.Float32)
	v += int(m.Float64)

	return v
}

func walkMessageData(m TestMessageData) int {
	var v int

	v += int(m.Int8())
	v += int(m.Int16())
	v += int(m.Int32())
	v += int(m.Int64())

	v += int(m.UInt8())
	v += int(m.UInt16())
	v += int(m.UInt32())
	v += int(m.UInt64())

	v += int(m.Float32())
	v += int(m.Float64())

	{
		list := m.List()
		for i := 0; i < list.Len(); i++ {
			v1 := list.Element(i).Int64()
			v += int(v1)
		}
	}

	{
		list := m.Messages()
		for i := 0; i < list.Len(); i++ {
			v1 := list.Element(i).Message()
			sub := TestSubMessageData{v1}

			v += int(sub.Int8())
			v += int(sub.Int16())
			v += int(sub.Int32())
			v += int(sub.Int64())
		}
	}

	{
		list := m.Strings()
		for i := 0; i < list.Len(); i++ {
			el := list.Element(i)
			s := el.String()
			v += len(s)
			if len(s) == 0 {
				panic("empty string")
			}
		}
	}

	return v
}

func compressedSize(b []byte) int {
	buf := &bytes.Buffer{}
	w := zlib.NewWriter(buf)
	w.Write(b)
	w.Close()
	c := buf.Bytes()
	return len(c)
}
