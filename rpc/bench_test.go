package rpc

import (
	"testing"
	"time"
)

func BenchmarkRequest(b *testing.B) {
	server := testEchoServer(b)

	conn := testConnect(b, server)
	defer conn.Free()

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	msg := "hello, world"

	for i := 0; i < b.N; i++ {
		req := testEchoRequest(b, msg)

		resp, st := conn.Request(nil, req)
		if !st.OK() {
			b.Fatal(st)
		}

		result := resp.Unwrap().Results().Get(0)
		msg1 := result.Value().String().Unwrap()
		if msg != msg1 {
			b.Fatal(msg1)
		}

		req.Free()
		resp.Free()
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

func BenchmarkRequest_Parallel(b *testing.B) {
	server := testEchoServer(b)

	conn := testConnect(b, server)
	defer conn.Free()

	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(10)
	t0 := time.Now()

	msg := "hello, world"

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			req := testEchoRequest(b, msg)

			resp, st := conn.Request(nil, req)
			if !st.OK() {
				b.Fatal(st)
			}

			result := resp.Unwrap().Results().Get(0)
			msg1 := result.Value().String().Unwrap()
			if msg != msg1 {
				b.Fatal(msg1)
			}

			req.Free()
			resp.Free()
		}
	})

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}
