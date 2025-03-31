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

func testClientConnect(t testing.TB, client *client) Conn {
	future, st := client.connect()
	if !st.OK() {
		t.Fatal(st)
	}

	select {
	case <-future.Wait():
	case <-time.After(time.Second):
		t.Fatal("connect timeout")
	}

	conn, st := future.Result()
	if !st.OK() {
		t.Fatal(st)
	}
	return conn
}

// Request

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
	opts := Default()
	opts.ClientMaxConns = 2

	server := testServerOpts(b, handle, opts)
	client := testClient(b, server)

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

			ch.Free()
			duration += time.Since(t0)
		}

		atomic.AddInt64(&totalDuration, int64(duration))
	})

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec
	conns := float64(client.conns.Load().len())
	latency := time.Duration(totalDuration) / time.Duration(b.N)

	b.ReportMetric(ops, "ops")
	b.ReportMetric(conns, "conns")
	b.ReportMetric(float64(latency.Microseconds())/1000, "latency,avg,ms")
}

// Stream

func BenchmarkClient_Stream_Parallel(b *testing.B) {
	closeMsg := []byte("close")
	handle := func(ctx Context, ch Channel) status.Status {
		for {
			msg, st := ch.Receive(ctx)
			if !st.OK() {
				return st
			}
			if !bytes.Equal(msg, closeMsg) {
				continue
			}

			break
		}

		return ch.Send(ctx, closeMsg)
	}

	server := testServer(b, handle)
	client := testClient(b, server)

	testClientConnect(b, client)
	testClientConnect(b, client)
	testClientConnect(b, client)
	testClientConnect(b, client)

	b.SetBytes(int64(benchMsgSize))
	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	b.RunParallel(func(p *testing.PB) {
		ctx := async.NoContext()

		ch, st := client.Channel(ctx)
		if !st.OK() {
			b.Fatal(st)
		}
		defer ch.Free()

		msg := bytes.Repeat([]byte("a"), benchMsgSize)

		for p.Next() {
			st = ch.Send(ctx, msg)
			if !st.OK() {
				b.Fatal(st)
			}
		}

		st = ch.Send(ctx, closeMsg)
		if !st.OK() {
			b.Fatal(st)
		}

		_, st = ch.Receive(ctx)
		if !st.OK() {
			b.Fatal(st)
		}
	})

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}
