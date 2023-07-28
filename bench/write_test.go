package spec

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/tests/pkg1"
)

func BenchmarkWrite_Small(b *testing.B) {
	obj := pkg1.TestSubobject(1)
	buf := buffer.NewSize(4096)

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()

	var size int
	for i := 0; i < b.N; i++ {
		buf.Reset()
		w := pkg1.NewSubmessageWriterBuffer(buf)

		data, err := obj.Write(w)
		if err != nil {
			b.Fatal(err)
		}

		size = len(data.Unwrap().Bytes())
		if size == 0 {
			b.Fatal()
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.SetBytes(int64(size))
	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
}

func BenchmarkWrite_Large(b *testing.B) {
	obj := pkg1.TestObject(b)
	buf := buffer.NewSize(4096)

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()

	var size int
	for i := 0; i < b.N; i++ {
		buf.Reset()
		w := pkg1.NewMessageWriterBuffer(buf)

		data, err := obj.Write(w)
		if err != nil {
			b.Fatal(err)
		}

		size = len(data.Unwrap().Bytes())
		if size == 0 {
			b.Fatal()
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.SetBytes(int64(size))
	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
}

// JSON

func BenchmarkJSON_Marshal_Small(b *testing.B) {
	obj := pkg1.TestSubobject(1)

	data, err := json.Marshal(obj)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.SetBytes(int64(len(data)))
	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(len(data)), "size")
	b.ReportMetric(float64(compressedSize(data)), "size-zlib")
}

func BenchmarkJSON_Marshal_Large(b *testing.B) {
	obj := pkg1.TestObject(b)

	data, err := json.Marshal(obj)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	t0 := time.Now()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.SetBytes(int64(len(data)))
	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(len(data)), "size")
	b.ReportMetric(float64(compressedSize(data)), "size-zlib")
}

func BenchmarkJSON_Encode_Large(b *testing.B) {
	obj := pkg1.TestObject(b)

	data, err := json.Marshal(obj)
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
		if err := e.Encode(obj); err != nil {
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

	b.SetBytes(int64(size))
	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(compressedSize(data)), "size-zlib")
}
