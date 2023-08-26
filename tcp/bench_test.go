package tcp

import (
	"bytes"
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
)

const msgSize = 16

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
	msg := bytes.Repeat([]byte("a"), msgSize)

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
		msg := bytes.Repeat([]byte("a"), msgSize)

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

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	b.RunParallel(func(p *testing.PB) {
		s, st := conn.Open(nil)
		if !st.OK() {
			b.Fatal(st)
		}
		defer s.Free()

		msg := bytes.Repeat([]byte("a"), msgSize)

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
