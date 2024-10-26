// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"bytes"
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
	t0 := time.Now()

	for i := 0; i < b.N; i++ {
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
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
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
	t0 := time.Now()

	b.RunParallel(func(p *testing.PB) {
		msg := bytes.Repeat([]byte("a"), benchMsgSize)

		for p.Next() {
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
		}
	})

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec
	conns := float64(len(client.conns))

	b.ReportMetric(ops, "ops")
	b.ReportMetric(conns, "conns")
}
