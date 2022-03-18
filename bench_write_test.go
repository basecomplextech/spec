package spec

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func Benchmark_Write(b *testing.B) {
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

		data, err := e.End()
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

// Standard JSON

func BenchmarkJSON_Write(b *testing.B) {
	msg := newTestObject()
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
}
