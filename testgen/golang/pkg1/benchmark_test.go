package pkg1

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/baseone-run/spec"
)

func BenchmarkRead(b *testing.B) {
	msg := testMessage()
	w := spec.NewWriter()
	if err := msg.Write(w); err != nil {
		b.Fatal(err)
	}
	data, err := w.End()
	if err != nil {
		b.Fatal(err)
	}

	size := len(data)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		if _, err := ReadMessage(data); err != nil {
			b.Fatal(err)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(size), "size")
}

func Benchmark_ReadData(b *testing.B) {
	msg := testMessage()

	w := spec.NewWriter()
	if err := msg.Write(w); err != nil {
		b.Fatal(err)
	}
	data, err := w.End()
	if err != nil {
		b.Fatal(err)
	}

	size := len(data)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		_, err := ReadMessageData(data)
		if err != nil {
			b.Fatal(err)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	rps := float64(b.N) / sec

	b.ReportMetric(rps, "rps")
	b.ReportMetric(float64(size), "size")
}

func Benchmark_JSONRead(b *testing.B) {
	msg := testMessage()

	data, err := json.Marshal(msg)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()

	msg1 := &Message{}
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
}
