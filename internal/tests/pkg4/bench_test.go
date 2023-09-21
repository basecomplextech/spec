package pkg4

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/logging"
)

func BenchmarkEmpty(b *testing.B) {
	logger := logging.TestLogger(b)
	service := newTestService()
	server := testServer(b, logger, service)
	client := testClient(b, logger, server)

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	for i := 0; i < b.N; i++ {
		st := client.Method(nil)
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
			st := client.Method(nil)
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
	logger := logging.TestLogger(b)
	service := newTestService()
	server := testServer(b, logger, service)
	client := testClient(b, logger, server)

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	for i := 0; i < b.N; i++ {
		_, _, _, st := client.Method2(nil, 1, 2, true)
		if !st.OK() {
			b.Fatal(st)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

func BenchmarkRequestFields_Parallel(b *testing.B) {
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
			_, _, _, st := client.Method2(nil, 1, 2, true)
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

// Request

func BenchmarkRequest(b *testing.B) {
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
		resp, st := client.Method3(nil, req)
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
			resp, st := client.Method3(nil, req)
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
