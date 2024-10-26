// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"bytes"
	"sync/atomic"
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/status"
)

func BenchmarkClient_Request(b *testing.B) {
	handle := func(ctx Context, ch Channel) status.Status {
		msg, st := ch.Receive(ctx)
		if !st.OK() {
			return st
		}
		return ch.SendAndClose(ctx, msg)
	}

	server := testServer(b, handle)
	client := testClient(b, server)

	ctx := async.NoContext()
	msg := bytes.Repeat([]byte("a"), benchMsgSize)

	b.ReportAllocs()
	b.ResetTimer()
	duration := time.Duration(0)

	for i := 0; i < b.N; i++ {
		t0 := time.Now()
		ch, st := client.Channel(ctx)
		if !st.OK() {
			b.Fatal(st)
		}

		st = ch.Send(ctx, msg)
		if !st.OK() {
			b.Fatal(st)
		}

		msg1, st := ch.Receive(ctx)
		if !st.OK() {
			b.Fatal(st)
		}
		if !bytes.Equal(msg, msg1) {
			b.Fatalf("expected %q, got %q", msg, msg1)
		}

		if st := ch.SendClose(ctx); !st.OK() {
			b.Fatal(st)
		}

		ch.Free()
		duration += time.Since(t0)
	}

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec
	latency := time.Duration(duration) / time.Duration(b.N)

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(latency.Microseconds())/1000, "latency,avg,ms")
}

func BenchmarkClient_Request_Parallel(b *testing.B) {
	handle := func(ctx Context, ch Channel) status.Status {
		msg, st := ch.Receive(ctx)
		if !st.OK() {
			return st
		}
		return ch.SendAndClose(ctx, msg)
	}

	ctx := async.NoContext()
	server := testServer(b, handle)
	client := testClient(b, server)
	client.options.ClientMaxConns = 4

	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(100)
	var totalDuration int64

	b.RunParallel(func(p *testing.PB) {
		msg := bytes.Repeat([]byte("a"), benchMsgSize)
		duration := time.Duration(0)

		for p.Next() {
			t0 := time.Now()
			ch, st := client.Channel(ctx)
			if !st.OK() {
				b.Fatal(st)
			}

			st = ch.Send(ctx, msg)
			if !st.OK() {
				b.Fatal(st)
			}

			msg1, st := ch.Receive(ctx)
			if !st.OK() {
				b.Fatal(st)
			}
			if !bytes.Equal(msg, msg1) {
				b.Fatalf("expected %q, got %q", msg, msg1)
			}

			if st := ch.SendClose(ctx); !st.OK() {
				b.Fatal(st)
			}

			ch.Free()
			duration += time.Since(t0)
		}

		atomic.AddInt64(&totalDuration, int64(duration))
	})

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec
	conns := float64(len(client.conns))
	latency := time.Duration(totalDuration) / time.Duration(b.N)

	b.ReportMetric(ops, "ops")
	b.ReportMetric(conns, "conns")
	b.ReportMetric(float64(latency.Microseconds())/1000, "latency,avg,ms")
}
