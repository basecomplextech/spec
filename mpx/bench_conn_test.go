// Copyright 2023 Ivan Korobkov. All rights reserved.
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

const benchMsgSize = 16

func BenchmarkRequest(b *testing.B) {
	handle := func(ctx Context, ch Channel) status.Status {
		msg, st := ch.Receive(ctx)
		if !st.OK() {
			return st
		}
		return ch.SendAndClose(ctx, msg)
	}

	server := testServer(b, handle)
	conn := testConnect(b, server)
	defer conn.Free()

	ctx := async.NoContext()
	msg := bytes.Repeat([]byte("a"), benchMsgSize)

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	for i := 0; i < b.N; i++ {
		ch, st := conn.Channel(ctx)
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
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

func BenchmarkRequest_Parallel(b *testing.B) {
	handle := func(ctx Context, ch Channel) status.Status {
		msg, st := ch.Receive(ctx)
		if !st.OK() {
			return st
		}
		return ch.SendAndClose(ctx, msg)
	}

	ctx := async.NoContext()
	server := testServer(b, handle)
	conn := testConnect(b, server)
	defer conn.Free()

	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(10)
	t0 := time.Now()

	b.RunParallel(func(p *testing.PB) {
		msg := bytes.Repeat([]byte("a"), benchMsgSize)

		for p.Next() {
			ch, st := conn.Channel(ctx)
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
		}
	})

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

// Stream

func BenchmarkStream(b *testing.B) {
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

		return ch.SendAndClose(ctx, closeMsg)
	}

	ctx := async.NoContext()
	server := testServer(b, handle)
	conn := testConnect(b, server)
	defer conn.Free()

	b.SetBytes(int64(benchMsgSize))
	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	ch, st := conn.Channel(ctx)
	if !st.OK() {
		b.Fatal(st)
	}
	defer ch.Free()

	msg := bytes.Repeat([]byte("a"), benchMsgSize)
	for i := 0; i < b.N; i++ {
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

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

func BenchmarkStream_16kb(b *testing.B) {
	close := []byte("close")
	benchMsgSize := 16 * 1024

	handle := func(ctx Context, ch Channel) status.Status {
		for {
			msg, st := ch.Receive(ctx)
			if !st.OK() {
				return st
			}
			if !bytes.Equal(msg, close) {
				continue
			}

			break
		}

		return ch.Send(ctx, close)
	}

	ctx := async.NoContext()
	server := testServer(b, handle)
	conn := testConnect(b, server)
	defer conn.Free()

	b.SetBytes(int64(benchMsgSize))
	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	ch, st := conn.Channel(ctx)
	if !st.OK() {
		b.Fatal(st)
	}
	defer ch.Free()

	msg := bytes.Repeat([]byte("a"), benchMsgSize)

	for i := 0; i < b.N; i++ {
		st = ch.Send(ctx, msg)
		if !st.OK() {
			b.Fatal(st)
		}
	}

	st = ch.Send(ctx, close)
	if !st.OK() {
		b.Fatal(st)
	}

	_, st = ch.Receive(ctx)
	if !st.OK() {
		b.Fatal(st)
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

func BenchmarkStream_Parallel(b *testing.B) {
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

	ctx := async.NoContext()
	server := testServer(b, handle)
	conn := testConnect(b, server)
	defer conn.Free()

	b.SetBytes(int64(benchMsgSize))
	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	b.RunParallel(func(p *testing.PB) {
		ch, st := conn.Channel(ctx)
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

func BenchmarkStream_16kb_Parallel(b *testing.B) {
	close := []byte("close")
	benchMsgSize := 16 * 1024

	handle := func(ctx Context, ch Channel) status.Status {
		for {
			msg, st := ch.Receive(ctx)
			if !st.OK() {
				return st
			}
			if !bytes.Equal(msg, close) {
				continue
			}

			break
		}

		return ch.Send(ctx, close)
	}

	ctx := async.NoContext()
	server := testServer(b, handle)
	conn := testConnect(b, server)
	defer conn.Free()

	b.SetBytes(int64(benchMsgSize))
	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	b.RunParallel(func(p *testing.PB) {
		ch, st := conn.Channel(ctx)
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

		st = ch.Send(ctx, close)
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

// OnClosed

func BenchmarkConn_OnClosed(b *testing.B) {
	handle := func(ctx Context, ch Channel) status.Status {
		msg, st := ch.Receive(ctx)
		if !st.OK() {
			return st
		}
		return ch.SendAndClose(ctx, msg)
	}

	server := testServer(b, handle)
	conn := testConnect(b, server)
	defer conn.Free()

	b.ReportAllocs()
	b.ResetTimer()

	fn := func() {}
	for i := 0; i < b.N; i++ {
		unsub := conn.OnClosed(fn)
		unsub()
	}

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops/1000_000, "mops")
}

func BenchmarkConn_OnClosed_Parallel(b *testing.B) {
	handle := func(ctx Context, ch Channel) status.Status {
		msg, st := ch.Receive(ctx)
		if !st.OK() {
			return st
		}
		return ch.SendAndClose(ctx, msg)
	}

	server := testServer(b, handle)
	conn := testConnect(b, server)
	defer conn.Free()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		fn := func() {}
		for p.Next() {
			unsub := conn.OnClosed(fn)
			unsub()
		}
	})

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops/1000_000, "mops")
}
