package pkg1

import (
	"testing"
	"time"

	"github.com/complexl/library/buffer"
	"github.com/complexl/spec"
)

func BenchmarkDecode(b *testing.B) {
	msg := testMessage(b)
	data := msg.RawBytes()
	size := len(data)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		_, _, err := DecodeMessage(data)
		if err != nil {
			b.Fatal(err)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
}

func BenchmarkEncode(b *testing.B) {
	buf := buffer.NewSize(1024)
	e := spec.NewEncoderBuffer(buf)

	msg := testMessage(b)
	data := msg.RawBytes()
	size := len(data)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		testEncode(b, e)

		buf.Reset()
		e.Init(buf)
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
}

func BenchmarkStructField(b *testing.B) {
	msg := testMessage(b)
	field := msg.msg.Field(52)
	size := len(field)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		msg.FieldStruct()
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
}
