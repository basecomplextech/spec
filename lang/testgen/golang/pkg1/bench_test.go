package pkg1

import (
	"testing"
	"time"

	"github.com/baseblck/library/buffer"
	spec "github.com/baseblck/spec/go"
)

func BenchmarkDecode(b *testing.B) {
	msg := testMessage(b)
	data := msg.Unwrap().Bytes()
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

	msg := testMessage(b)
	data := msg.Unwrap().Bytes()
	size := len(data)

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		e := spec.NewEncoderBuffer(buf)
		testEncode(b, e)
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
