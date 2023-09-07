package rpc

import (
	"testing"
	"time"
)

func BenchmarkRequest(b *testing.B) {
	server := testEchoServer(b)

	client := testClient(b, server)
	defer client.Close()

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	msg := "hello, world"
	req := testEchoRequest(b, msg)

	for i := 0; i < b.N; i++ {
		result, st := client.Request(nil, req)
		if !st.OK() {
			b.Fatal(st)
		}

		result.Free()
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

func BenchmarkRequest_Parallel(b *testing.B) {
	server := testEchoServer(b)

	client := testClient(b, server)
	defer client.Close()

	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(10)
	t0 := time.Now()

	msg := "hello, world"

	b.RunParallel(func(p *testing.PB) {
		req := testEchoRequest(b, msg)

		for p.Next() {
			result, st := client.Request(nil, req)
			if !st.OK() {
				b.Fatal(st)
			}

			result.Free()
		}
	})

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}
