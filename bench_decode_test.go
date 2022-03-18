package spec

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"testing"
	"time"
)

func BenchmarkDecode(b *testing.B) {
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
		_, _, err := DecodeTestMessage(data)
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

func BenchmarkDecodeObject(b *testing.B) {
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
		if err := msg.Decode(data); err != nil {
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

// Standard JSON

func BenchmarkJSON_Unmarshal(b *testing.B) {
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

func compressedSize(b []byte) int {
	buf := &bytes.Buffer{}
	e := zlib.NewWriter(buf)
	e.Write(b)
	e.Close()
	c := buf.Bytes()
	return len(c)
}
