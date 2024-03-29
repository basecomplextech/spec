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
	handle := func(ctx async.Context, ch Channel) status.Status {
		msg, st := ch.ReadSync(ctx)
		if !st.OK() {
			return st
		}
		return ch.WriteAndClose(ctx, msg)
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

		st = ch.Write(ctx, msg)
		if !st.OK() {
			b.Fatal(st)
		}

		msg1, st := ch.ReadSync(ctx)
		if !st.OK() {
			b.Fatal(st)
		}
		if !bytes.Equal(msg, msg1) {
			b.Fatalf("expected %q, got %q", msg, msg1)
		}

		if st := ch.Close(); !st.OK() {
			b.Fatal(st)
		}

		ch.Free()
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

func BenchmarkRequest_Parallel(b *testing.B) {
	handle := func(ctx async.Context, ch Channel) status.Status {
		msg, st := ch.ReadSync(ctx)
		if !st.OK() {
			return st
		}
		return ch.WriteAndClose(ctx, msg)
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

			st = ch.Write(ctx, msg)
			if !st.OK() {
				b.Fatal(st)
			}

			msg1, st := ch.ReadSync(ctx)
			if !st.OK() {
				b.Fatal(st)
			}
			if !bytes.Equal(msg, msg1) {
				b.Fatalf("expected %q, got %q", msg, msg1)
			}

			if st := ch.Close(); !st.OK() {
				b.Fatal(st)
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
	handle := func(ctx async.Context, ch Channel) status.Status {
		for {
			msg, st := ch.ReadSync(ctx)
			if !st.OK() {
				return st
			}
			if !bytes.Equal(msg, closeMsg) {
				continue
			}

			break
		}

		st := ch.Write(ctx, closeMsg)
		if !st.OK() {
			return st
		}
		return ch.Close()
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
		st = ch.Write(ctx, msg)
		if !st.OK() {
			b.Fatal(st)
		}
	}

	st = ch.Write(ctx, closeMsg)
	if !st.OK() {
		b.Fatal(st)
	}

	_, st = ch.ReadSync(ctx)
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

	handle := func(ctx async.Context, ch Channel) status.Status {
		for {
			msg, st := ch.ReadSync(ctx)
			if !st.OK() {
				return st
			}
			if !bytes.Equal(msg, close) {
				continue
			}

			break
		}

		st := ch.Write(ctx, close)
		if !st.OK() {
			return st
		}
		return ch.Close()
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
		st = ch.Write(ctx, msg)
		if !st.OK() {
			b.Fatal(st)
		}
	}

	st = ch.Write(ctx, close)
	if !st.OK() {
		b.Fatal(st)
	}

	_, st = ch.ReadSync(ctx)
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
	handle := func(ctx async.Context, ch Channel) status.Status {
		for {
			msg, st := ch.ReadSync(ctx)
			if !st.OK() {
				return st
			}
			if !bytes.Equal(msg, closeMsg) {
				continue
			}

			break
		}

		st := ch.Write(ctx, closeMsg)
		if !st.OK() {
			return st
		}
		return ch.Close()
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
			st = ch.Write(ctx, msg)
			if !st.OK() {
				b.Fatal(st)
			}
		}

		st = ch.Write(ctx, closeMsg)
		if !st.OK() {
			b.Fatal(st)
		}

		_, st = ch.ReadSync(ctx)
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

	handle := func(ctx async.Context, ch Channel) status.Status {
		for {
			msg, st := ch.ReadSync(ctx)
			if !st.OK() {
				return st
			}
			if !bytes.Equal(msg, close) {
				continue
			}

			break
		}

		st := ch.Write(ctx, close)
		if !st.OK() {
			return st
		}
		return ch.Close()
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
			st = ch.Write(ctx, msg)
			if !st.OK() {
				b.Fatal(st)
			}
		}

		st = ch.Write(ctx, close)
		if !st.OK() {
			b.Fatal(st)
		}

		_, st = ch.ReadSync(ctx)
		if !st.OK() {
			b.Fatal(st)
		}
	})

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}
