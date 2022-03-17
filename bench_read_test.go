package spec

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"testing"
	"time"
)

func Benchmark_Read(b *testing.B) {
	msg := newTestObject()
	data, err := msg.Marshal()
	if err != nil {
		b.Fatal(err)
	}

	size := len(data)
	compressed := compressedSize(data)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		_, _, err := ReadTestMessage(data)
		if err != nil {
			b.Fatal(err)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(compressed), "size-zlib")
}

func Benchmark_Walk(b *testing.B) {
	msg := newTestObject()
	data, err := msg.Marshal()
	if err != nil {
		b.Fatal(err)
	}

	size := len(data)
	compressed := compressedSize(data)

	d, _, err := ReadTestMessage(data)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		v, err := walkMessage(d)
		if err != nil {
			b.Fatal(err)
		}
		if v == 0 {
			b.Fatal()
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(compressed), "size-zlib")
}

func Benchmark_ReadObject(b *testing.B) {
	msg := newTestObject()
	data, err := msg.Marshal()
	if err != nil {
		b.Fatal(err)
	}

	size := len(data)
	compressed := compressedSize(data)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		if err := msg.Read(data); err != nil {
			b.Fatal(err)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(compressed), "size-zlib")
}

func Benchmark_WriteObject(b *testing.B) {
	msg := newTestObject()
	buf := make([]byte, 0, 4096)
	w := NewWriterBuffer(buf)

	data, err := msg.Marshal()
	if err != nil {
		b.Fatal(err)
	}

	size := len(data)
	compressed := compressedSize(data)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		mw := WriteTestMessage(w)
		if err := msg.Write(mw); err != nil {
			b.Fatal(err)
		}
		if _, err := w.End(); err != nil {
			b.Fatal(err)
		}

		w.Reset()
		w.buf = buf[:0]
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(compressed), "size-zlib")
}

// Standard JSON

func Benchmark_JSONMarshal(b *testing.B) {
	msg := newTestObject()
	data, err := json.Marshal(msg)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(len(data)), "size")
	b.ReportMetric(float64(compressedSize(data)), "size-zlib")
}

func Benchmark_JSONUnmarshal(b *testing.B) {
	msg := newTestObject()

	data, err := json.Marshal(msg)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()

	msg1 := &TestObject{}
	for i := 0; i < b.N; i++ {
		if err := json.Unmarshal(data, msg1); err != nil {
			b.Fatal(err)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(len(data)), "size")
	b.ReportMetric(float64(compressedSize(data)), "size-zlib")
}

func Benchmark_JSONEncode(b *testing.B) {
	msg := newTestObject()
	data, err := json.Marshal(msg)
	if err != nil {
		b.Fatal(err)
	}

	buf := &bytes.Buffer{}

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		buf.Reset()

		if err := json.NewEncoder(buf).Encode(msg); err != nil {
			b.Fatal(err)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(len(data)), "size")
	b.ReportMetric(float64(compressedSize(data)), "size-zlib")
}

// private

func walkMessage(m TestMessage) (int, error) {
	var v int

	v += int(m.Byte())

	v += int(m.Int32())
	v += int(m.Int64())

	v += int(m.Uint32())
	v += int(m.Uint64())

	v += int(m.Float32())
	v += int(m.Float64())

	v += len(m.String())
	v += len(m.Bytes())

	{
		list := m.List()
		for i := 0; i < list.Count(); i++ {
			v1 := list.Element(i)
			v += int(v1)
		}
	}

	{
		list := m.Messages()
		for i := 0; i < list.Count(); i++ {
			sub := list.Element(i)

			v += int(sub.Byte())
			v += int(sub.Int32())
			v += int(sub.Int64())
		}
	}

	{
		list := m.Strings()
		for i := 0; i < list.Count(); i++ {
			v := list.Element(i)
			if len(v) == 0 {
				panic("empty string")
			}
		}
	}

	{
		str := m.Struct()
		v += int(str.X)
		v += int(str.Y)
	}
	return v, nil
}

func compressedSize(b []byte) int {
	buf := &bytes.Buffer{}
	w := zlib.NewWriter(buf)
	w.Write(b)
	w.Close()
	c := buf.Bytes()
	return len(c)
}
