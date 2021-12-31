package spec

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"testing"
	"time"
)

func Benchmark_Marshal(b *testing.B) {
	msg := newTestMessage()
	data, err := msg.Marshal()
	if err != nil {
		b.Fatal(err)
	}

	size := len(data)
	compressed := compressedSize(data)
	buf := make([]byte, 0, 4096)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		buf, err = msg.MarshalTo(buf)
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

func Benchmark_Unmarshal(b *testing.B) {
	msg := newTestMessage()
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
		if err := msg.Unmarshal(data); err != nil {
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

func Benchmark_ReadData(b *testing.B) {
	msg := newTestMessage()
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
		_, err := readTestMessageData(data)
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
	msg := newTestMessage()
	data, err := msg.Marshal()
	if err != nil {
		b.Fatal(err)
	}

	size := len(data)
	compressed := compressedSize(data)

	d, err := readTestMessageData(data)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		v, err := walkMessageData(d)
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

// Standard JSON

func Benchmark_JSONMarshal(b *testing.B) {
	msg := newTestMessage()
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
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(len(data)), "size")
	b.ReportMetric(float64(compressedSize(data)), "size-zlib")
}

func Benchmark_JSONEncode(b *testing.B) {
	msg := newTestMessage()
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

func walkMessageData(m TestMessageData) (int, error) {
	var v int

	v += int(m.Int8())
	v += int(m.Int16())
	v += int(m.Int32())
	v += int(m.Int64())

	v += int(m.Uint8())
	v += int(m.Uint16())
	v += int(m.Uint32())
	v += int(m.Uint64())

	v += int(m.Float32())
	v += int(m.Float64())

	v += len(m.String())
	v += len(m.Bytes())

	{
		list := m.List()
		for i := 0; i < list.Len(); i++ {
			v1 := list.Int64(i)
			v += int(v1)
		}
	}

	{
		list := m.Messages()
		for i := 0; i < list.Len(); i++ {
			data := list.Element(i)
			if len(data) == 0 {
				continue
			}

			sub, err := getTestSubMessageData(data)
			if err != nil {
				return 0, err
			}

			v += int(sub.Int8())
			v += int(sub.Int16())
			v += int(sub.Int32())
			v += int(sub.Int64())
		}
	}

	{
		list := m.Strings()
		for i := 0; i < list.Len(); i++ {
			v := list.String(i)
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
