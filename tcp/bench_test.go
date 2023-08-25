package tcp

import (
	"bytes"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
)

// Open/close

func BenchmarkOpenClose(b *testing.B) {
	handle := func(stream Stream) status.Status {
		for {
			msg, st := stream.Read(nil)
			if !st.OK() {
				return st
			}
			if st := stream.Write(nil, msg); !st.OK() {
				return st
			}
		}
	}

	logger := logging.TestLogger(b)
	server := testServer(b, "localhost:0", handle, logger)

	run, st := server.Run()
	if !st.OK() {
		b.Fatal(st)
	}
	defer async.CancelWait(run)

	select {
	case <-server.listening.Wait():
	case <-time.After(time.Second):
		b.Fatal("server not listening")
	}

	addr := server.listenAddress()
	conn, st := Dial(addr, logger)
	if !st.OK() {
		b.Fatal(st)
	}
	defer conn.Free()
	msg := bytes.Repeat([]byte("a"), 16)

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	for i := 0; i < b.N; i++ {
		s, st := conn.Open(nil)
		if !st.OK() {
			b.Fatal(st)
		}

		st = s.Write(nil, msg)
		if !st.OK() {
			b.Fatal(st)
		}

		msg1, st := s.Read(nil)
		if !st.OK() {
			b.Fatal(st)
		}
		if !bytes.Equal(msg, msg1) {
			b.Fatalf("expected %q, got %q", msg, msg1)
		}

		if st := s.Close(nil); !st.OK() {
			b.Fatal(st)
		}

		s.Free()
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

func BenchmarkOpenClose_Parallel(b *testing.B) {
	handle := func(stream Stream) status.Status {
		for {
			msg, st := stream.Read(nil)
			if !st.OK() {
				return st
			}
			if st := stream.Write(nil, msg); !st.OK() {
				return st
			}
		}
	}

	logger := logging.TestLogger(b)
	server := testServer(b, "localhost:0", handle, logger)

	run, st := server.Run()
	if !st.OK() {
		b.Fatal(st)
	}
	defer async.CancelWait(run)

	select {
	case <-server.listening.Wait():
	case <-time.After(time.Second):
		b.Fatal("server not listening")
	}

	addr := server.listenAddress()
	conn, st := Dial(addr, logger)
	if !st.OK() {
		b.Fatal(st)
	}
	defer conn.Free()

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	b.RunParallel(func(p *testing.PB) {
		msg := bytes.Repeat([]byte("a"), 16)

		for p.Next() {
			s, st := conn.Open(nil)
			if !st.OK() {
				b.Fatal(st)
			}

			st = s.Write(nil, msg)
			if !st.OK() {
				b.Fatal(st)
			}

			msg1, st := s.Read(nil)
			if !st.OK() {
				b.Fatal(st)
			}

			if !bytes.Equal(msg, msg1) {
				b.Fatalf("expected %q, got %q", msg, msg1)
			}

			if st := s.Close(nil); !st.OK() {
				b.Fatal(st)
			}

			s.Free()
		}
	})

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

// Request/reply

func BenchmarkRequestReply(b *testing.B) {
	handle := func(stream Stream) status.Status {
		for {
			msg, st := stream.Read(nil)
			if !st.OK() {
				return st
			}
			if st := stream.Write(nil, msg); !st.OK() {
				return st
			}
		}
	}

	logger := logging.TestLogger(b)
	server := testServer(b, "localhost:0", handle, logger)

	run, st := server.Run()
	if !st.OK() {
		b.Fatal(st)
	}
	defer async.CancelWait(run)

	select {
	case <-server.listening.Wait():
	case <-time.After(time.Second):
		b.Fatal("server not listening")
	}

	addr := server.listenAddress()
	conn, st := Dial(addr, logger)
	if !st.OK() {
		b.Fatal(st)
	}
	defer conn.Free()

	stream, st := conn.Open(nil)
	if !st.OK() {
		b.Fatal(st)
	}
	msg := bytes.Repeat([]byte("a"), 16)

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	for i := 0; i < b.N; i++ {
		st = stream.Write(nil, msg)
		if !st.OK() {
			b.Fatal(st)
		}

		msg1, st := stream.Read(nil)
		if !st.OK() {
			b.Fatal(st)
		}

		if !bytes.Equal(msg, msg1) {
			b.Fatalf("expected %q, got %q", msg, msg1)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

func BenchmarkRequestReply_Parallel(b *testing.B) {
	handle := func(stream Stream) status.Status {
		for {
			msg, st := stream.Read(nil)
			if !st.OK() {
				return st
			}
			if st := stream.Write(nil, msg); !st.OK() {
				return st
			}
		}
	}

	logger := logging.TestLogger(b)
	server := testServer(b, "localhost:0", handle, logger)

	run, st := server.Run()
	if !st.OK() {
		b.Fatal(st)
	}
	defer async.CancelWait(run)

	select {
	case <-server.listening.Wait():
	case <-time.After(time.Second):
		b.Fatal("server not listening")
	}

	addr := server.listenAddress()
	conn, st := Dial(addr, logger)
	if !st.OK() {
		b.Fatal(st)
	}
	defer conn.Free()

	cpu := runtime.GOMAXPROCS(0)
	streams := make([]Stream, 0, cpu)

	for i := 0; i < cpu; i++ {
		stream, st := conn.Open(nil)
		if !st.OK() {
			b.Fatal(st)
		}
		streams = append(streams, stream)
	}

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	var i int64
	b.RunParallel(func(p *testing.PB) {
		j := atomic.AddInt64(&i, 1) - 1
		s := streams[j]
		msg := bytes.Repeat([]byte("a"), 16)

		for p.Next() {
			st = s.Write(nil, msg)
			if !st.OK() {
				b.Fatal(st)
			}

			msg1, st := s.Read(nil)
			if !st.OK() {
				b.Fatal(st)
			}

			if !bytes.Equal(msg, msg1) {
				b.Fatalf("expected %q, got %q", msg, msg1)
			}
		}
	})

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

// Stream

func BenchmarkStream_Parallel(b *testing.B) {
	handle := func(stream Stream) status.Status {
		for {
			_, st := stream.Read(nil)
			if !st.OK() {
				return st
			}
		}
	}

	logger := logging.TestLogger(b)
	server := testServer(b, "localhost:0", handle, logger)

	run, st := server.Run()
	if !st.OK() {
		b.Fatal(st)
	}
	defer async.CancelWait(run)

	select {
	case <-server.listening.Wait():
	case <-time.After(time.Second):
		b.Fatal("server not listening")
	}

	addr := server.listenAddress()
	conn, st := Dial(addr, logger)
	if !st.OK() {
		b.Fatal(st)
	}
	defer conn.Free()

	cpu := runtime.GOMAXPROCS(0)
	streams := make([]Stream, 0, cpu)

	for i := 0; i < cpu; i++ {
		stream, st := conn.Open(nil)
		if !st.OK() {
			b.Fatal(st)
		}
		streams = append(streams, stream)
	}

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	var i int64
	b.RunParallel(func(p *testing.PB) {
		j := atomic.AddInt64(&i, 1) - 1
		s := streams[j]
		msg := bytes.Repeat([]byte("a"), 16)

		for p.Next() {
			st = s.Write(nil, msg)
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
