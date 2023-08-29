package tcp

import (
	"bytes"
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/status"
)

const benchMsgSize = 16

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

	server := testServer(b, handle)
	conn := testConnect(b, server)
	defer conn.Free()

	msg := bytes.Repeat([]byte("a"), benchMsgSize)

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

		if st := s.Close(); !st.OK() {
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

			if st := s.Close(); !st.OK() {
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
	close := []byte("close")
	handle := func(s Stream) status.Status {
		for {
			msg, st := s.Read(nil)
			if !st.OK() {
				return st
			}
			if !bytes.Equal(msg, close) {
				continue
			}

			break
		}

		st := s.Write(nil, close)
		if !st.OK() {
			return st
		}
		return s.Close()
	}

	server := testServer(b, handle)
	conn := testConnect(b, server)
	defer conn.Free()

	b.SetBytes(int64(benchMsgSize))
	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	b.RunParallel(func(p *testing.PB) {
		s, st := conn.Open(nil)
		if !st.OK() {
			b.Fatal(st)
		}
		defer s.Free()

		msg := bytes.Repeat([]byte("a"), benchMsgSize)

		for p.Next() {
			st = s.Write(nil, msg)
			if !st.OK() {
				b.Fatal(st)
			}
		}

		st = s.Write(nil, close)
		if !st.OK() {
			b.Fatal(st)
		}

		_, st = s.Read(nil)
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

	handle := func(s Stream) status.Status {
		for {
			msg, st := s.Read(nil)
			if !st.OK() {
				return st
			}
			if !bytes.Equal(msg, close) {
				continue
			}

			break
		}

		st := s.Write(nil, close)
		if !st.OK() {
			return st
		}
		return s.Close()
	}

	server := testServer(b, handle)
	conn := testConnect(b, server)
	defer conn.Free()

	b.SetBytes(int64(benchMsgSize))
	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	b.RunParallel(func(p *testing.PB) {
		s, st := conn.Open(nil)
		if !st.OK() {
			b.Fatal(st)
		}
		defer s.Free()

		msg := bytes.Repeat([]byte("a"), benchMsgSize)

		for p.Next() {
			st = s.Write(nil, msg)
			if !st.OK() {
				b.Fatal(st)
			}
		}

		st = s.Write(nil, close)
		if !st.OK() {
			b.Fatal(st)
		}

		_, st = s.Read(nil)
		if !st.OK() {
			b.Fatal(st)
		}
	})

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}
