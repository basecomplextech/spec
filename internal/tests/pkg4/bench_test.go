// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package pkg4

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
)

func BenchmarkEmpty(b *testing.B) {
	ctx := async.NoContext()
	logger := logging.TestLogger(b)

	service := newTestService()
	server := testServer(b, logger, service)
	client := testClient(b, logger, server)

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	for i := 0; i < b.N; i++ {
		st := client.Method(ctx)
		if !st.OK() {
			b.Fatal(st)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

func BenchmarkEmpty_Parallel(b *testing.B) {
	ctx := async.NoContext()
	logger := logging.TestLogger(b)

	service := newTestService()
	server := testServer(b, logger, service)
	client := testClient(b, logger, server)

	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(10)
	t0 := time.Now()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			st := client.Method(ctx)
			if !st.OK() {
				b.Fatal(st)
			}
		}
	})

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

// Request fields

func BenchmarkRequestFields(b *testing.B) {
	ctx := async.NoContext()
	logger := logging.TestLogger(b)

	service := newTestService()
	server := testServer(b, logger, service)
	client := testClient(b, logger, server)

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	buf := alloc.NewBuffer()

	for i := 0; i < b.N; i++ {
		buf.Reset()

		w := NewServiceMethod2RequestWriterBuffer(buf)
		w.A(1)
		w.B(2)
		w.C(true)

		req, err := w.Build()
		if err != nil {
			b.Fatal(err)
		}

		resp, st := client.Method2(ctx, req)
		if !st.OK() {
			b.Fatal(st)
		}
		resp.Release()
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

func BenchmarkRequestFields_Parallel(b *testing.B) {
	ctx := async.NoContext()
	logger := logging.TestLogger(b)

	service := newTestService()
	server := testServer(b, logger, service)
	client := testClient(b, logger, server)

	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(10)
	t0 := time.Now()

	b.RunParallel(func(p *testing.PB) {
		buf := alloc.NewBuffer()

		for p.Next() {
			buf.Reset()

			w := NewServiceMethod2RequestWriterBuffer(buf)
			w.A(1)
			w.B(2)
			w.C(true)

			req, err := w.Build()
			if err != nil {
				b.Fatal(err)
			}

			resp, st := client.Method2(ctx, req)
			if !st.OK() {
				b.Fatal(st)
			}
			resp.Release()
		}
	})

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

// Request

func BenchmarkRequest(b *testing.B) {
	ctx := async.NoContext()
	logger := logging.TestLogger(b)

	service := newTestService()
	server := testServer(b, logger, service)
	client := testClient(b, logger, server)

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	w := NewRequestWriter()
	w.Msg("hello")
	req, err := w.Build()
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		resp, st := client.Method3(ctx, req)
		if !st.OK() {
			b.Fatal(st)
		}
		resp.Release()
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

func BenchmarkRequest_Parallel(b *testing.B) {
	ctx := async.NoContext()
	logger := logging.TestLogger(b)

	service := newTestService()
	server := testServer(b, logger, service)
	client := testClient(b, logger, server)

	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(10)
	t0 := time.Now()

	b.RunParallel(func(p *testing.PB) {
		w := NewRequestWriter()
		w.Msg("hello")
		req, err := w.Build()
		if err != nil {
			b.Fatal(err)
		}

		for p.Next() {
			resp, st := client.Method3(ctx, req)
			if !st.OK() {
				b.Fatal(st)
			}
			resp.Release()
		}
	})

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

// Oneway

func BenchmarkOneway(b *testing.B) {
	ctx := async.NoContext()
	logger := logging.TestLogger(b)

	service := newTestService()
	server := testServer(b, logger, service)
	client := testClient(b, logger, server)

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	w := NewServiceMethod0RequestWriter()
	w.Msg("hello")
	req, err := w.Build()
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		st := client.Method0(ctx, req)
		if !st.OK() {
			b.Fatal(st)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}
