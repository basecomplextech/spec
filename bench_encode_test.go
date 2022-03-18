package spec

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func BenchmarkEncode(b *testing.B) {
	msg := newTestObject()
	buf := make([]byte, 0, 4096)
	e := NewEncoderBuffer(buf)

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()

	var size int
	for i := 0; i < b.N; i++ {
		me := BeginTestMessage(e)
		if err := msg.Encode(me); err != nil {
			b.Fatal(err)
		}

		data, err := me.End()
		if err != nil {
			b.Fatal(err)
		}
		if len(data) < 100 {
			b.Fatal(len(data))
		}

		// b.Fatal(len(data))
		e.Reset()
		e.buf = buf[:0]

		size = len(data)
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(size), "size")
}

func BenchmarkEncodeObject(b *testing.B) {
	msg := newTestObject()
	buf := make([]byte, 0, 4096)
	e := NewEncoderBuffer(buf)

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
		me := BeginTestMessage(e)
		if err := msg.Encode(me); err != nil {
			b.Fatal(err)
		}
		if _, err := e.End(); err != nil {
			b.Fatal(err)
		}

		e.Reset()
		e.buf = buf[:0]
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(compressed), "size-zlib")
}

// Standard JSON

func BenchmarkJSON_Marshal(b *testing.B) {
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

func BenchmarkJSON_Encode(b *testing.B) {
	msg := newTestObject()
	data, err := json.Marshal(msg)
	if err != nil {
		b.Fatal(err)
	}

	buf := make([]byte, 0, 4096)
	buffer := bytes.NewBuffer(buf)

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()

	var size int
	for i := 0; i < b.N; i++ {
		e := json.NewEncoder(buffer)
		if err := e.Encode(msg); err != nil {
			b.Fatal(err)
		}
		if buffer.Len() < 100 {
			b.Fatal(buffer.Len())
		}

		size = buffer.Len()
		buffer.Reset()
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(compressedSize(data)), "size-zlib")
}
