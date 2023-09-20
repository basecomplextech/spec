package rpc

import (
	"bytes"
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/stretchr/testify/assert"
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

		result.Release()
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

			result.Release()
		}
	})

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

// Stream

func BenchmarkStream(b *testing.B) {
	streamMsg := []byte("hello, world")
	closeMsg := []byte("close")

	handle := func(cancel <-chan struct{}, ch ServerChannel) (*alloc.Buffer, status.Status) {
		for {
			msg, st := ch.Receive(cancel)
			if !st.OK() {
				return nil, st
			}
			if bytes.Equal(msg, closeMsg) {
				break
			}
		}

		buf := alloc.NewBuffer()
		w := spec.NewValueWriterBuffer(buf)
		w.String("response")
		if _, err := w.Build(); err != nil {
			return nil, status.WrapError(err)
		}
		return buf, status.OK
	}

	server := testServer(b, handle)
	client := testClient(b, server)
	defer client.Close()

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	{
		req := testEchoRequest(b, "request")
		ch, st := client.Channel(nil, req)
		if !st.OK() {
			b.Fatal(st)
		}
		defer ch.Free()

		for i := 0; i < b.N; i++ {
			st = ch.Send(nil, streamMsg)
			if !st.OK() {
				b.Fatal(st)
			}
		}

		st = ch.Send(nil, []byte(closeMsg))
		if !st.OK() {
			b.Fatal(st)
		}

		result, st := ch.Response(nil)
		if !st.OK() {
			b.Fatal(st)
		}
		defer result.Release()

		assert.Equal(b, "response", result.Unwrap().String().Unwrap())
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}
