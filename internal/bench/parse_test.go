// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package spec

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"testing"

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

	for i := 0; i < b.N; i++ {
		_, _, err := pkg1.ParseMessage(bytes)
		if err != nil {
			b.Fatal(err)
		}
	}

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops/1000_000, "mops")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(compressed), "size-zlib")
}

// ReadMessage

func BenchmarkReadMessage(b *testing.B) {
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

	for i := 0; i < b.N; i++ {
		m, err := pkg1.NewMessageErr(bytes)
		if err != nil {
			b.Fatal(err)
		}

		_ = m.Bool()
		_ = m.Byte()

		_ = m.Int16()
		_ = m.Int32()
		_ = m.Int64()

		_ = m.Uint16()
		_ = m.Uint32()
		_ = m.Uint64()

		_ = m.Float32()
		_ = m.Float64()

		_ = m.Bin64()
		_ = m.Bin128()
		_ = m.Bin256()

		_ = m.String()
		_ = m.Bytes1()

		_ = m.Message1()
		_ = m.Enum1()
		_ = m.Struct1()
		_ = m.Submessage()
		_ = m.Submessage1()

		_ = m.Ints()
		_ = m.Strings()
		_ = m.Structs()
		_ = m.Submessages()
		_ = m.Submessages1()

		_ = m.Any()
	}

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops/1000_000, "mops")
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

	for i := 0; i < b.N; i++ {
		msg := pkg1.NewMessage(bytes)
		if msg.Unwrap().Len() == 0 {
			b.Fatal()
		}
	}

	sec := b.Elapsed().Seconds()
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

	msg1 := &pkg1.Object{}
	for i := 0; i < b.N; i++ {
		if err := json.Unmarshal(data, msg1); err != nil {
			b.Fatal(err)
		}
	}

	sec := b.Elapsed().Seconds()
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
