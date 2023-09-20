package pkg4

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/logging"
)

func BenchmarkRequest(b *testing.B) {
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
