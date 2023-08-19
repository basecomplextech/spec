package spec

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"testing"
	"time"

	"github.com/basecomplextech/spec/internal/tests/pkg1"
)

func BenchmarkParseMessage(b *testing.B) {
	obj := pkg1.TestObject(b)
	msg, err := obj.Write(pkg1.NewMessageWriter())
	if err != nil {
		b.Fatal(err)
	}

	bytes := msg.Unwrap().Raw()
	size := len(bytes)
	compressed := compressedSize(bytes)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		_, _, err := pkg1.ParseMessage(bytes)
		if err != nil {
			b.Fatal(err)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(compressed), "size-zlib")
}

// NewMessage

func BenchmarkNewMessage(b *testing.B) {
	obj := pkg1.TestObject(b)
	msg, err := obj.Write(pkg1.NewMessageWriter())
	if err != nil {
		b.Fatal(err)
	}

	bytes := msg.Unwrap().Raw()
	size := len(bytes)
	compressed := compressedSize(bytes)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		msg := pkg1.NewMessage(bytes)
		if len(msg.Unwrap().Raw()) == 0 {
			b.Fatal()
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(compressed), "size-zlib")
}

// JSON

func BenchmarkJSON_Unmarshal(b *testing.B) {
	obj := pkg1.TestObject(b)

	data, err := json.Marshal(obj)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()

	msg1 := &pkg1.Object{}
	for i := 0; i < b.N; i++ {
		if err := json.Unmarshal(data, msg1); err != nil {
			b.Fatal(err)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(len(data)), "size")
	b.ReportMetric(float64(compressedSize(data)), "size-zlib")
}

// util

func compressedSize(b []byte) int {
	buf := &bytes.Buffer{}
	e := zlib.NewWriter(buf)
	e.Write(b)
	e.Close()
	c := buf.Bytes()
	return len(c)
}
