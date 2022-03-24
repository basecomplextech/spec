package spec

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/complexl/library/buffer"
)

func BenchmarkEncode_Small(b *testing.B) {
	msg := newTestSmall()
	buf := buffer.NewSize(4096)

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()

	var size int
	for i := 0; i < b.N; i++ {
		buf.Reset()
		e := NewEncoderBuffer(buf)

		data, err := msg.Encode(e)
		if err != nil {
			b.Fatal(err)
		}
		if len(data) == 0 {
			b.Fatal()
		}

		size = len(data)
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
	b.SetBytes(int64(size))
}

func BenchmarkEncode_Large(b *testing.B) {
	msg := newTestObject()
	buf := buffer.NewSize(4096)

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()

	var size int
	for i := 0; i < b.N; i++ {
		buf.Reset()
		e := NewEncoderBuffer(buf)

		builder, err := BuildTestMessageEncoder(e)
		if err != nil {
			b.Fatal(err)
		}
		if err := msg.Encode(builder); err != nil {
			b.Fatal(err)
		}

		data, err := builder.End()
		if err != nil {
			b.Fatal(err)
		}
		if len(data) < 100 {
			b.Fatal(len(data))
		}

		size = len(data)
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
	b.SetBytes(int64(size))
}

func BenchmarkEncodeObject(b *testing.B) {
	msg := newTestObject()
	buf := buffer.NewSize(4096)
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
		builder, err := BuildTestMessageEncoder(e)
		if err != nil {
			b.Fatal(err)
		}
		if err := msg.Encode(builder); err != nil {
			b.Fatal(err)
		}
		if _, err := e.End(); err != nil {
			b.Fatal(err)
		}

		buf.Reset()
		e.init(buf)
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(compressed), "size-zlib")
}

// Standard JSON

func BenchmarkJSON_Marshal_Small(b *testing.B) {
	msg := newTestSmall()
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
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(len(data)), "size")
	b.ReportMetric(float64(compressedSize(data)), "size-zlib")
}

func BenchmarkJSON_Marshal_Large(b *testing.B) {
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
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(len(data)), "size")
	b.ReportMetric(float64(compressedSize(data)), "size-zlib")
}

func BenchmarkJSON_Encode_Large(b *testing.B) {
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
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(compressedSize(data)), "size-zlib")
}
