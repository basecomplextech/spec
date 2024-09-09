// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package rpc

import (
	"bytes"
	"testing"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/ref"
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

	ctx := async.NoContext()
	msg := "hello, world"
	req := testEchoRequest(b, msg)

	for i := 0; i < b.N; i++ {
		result, st := client.Request(ctx, req)
		if !st.OK() {
			b.Fatal(st)
		}

		result.Release()
	}

	sec := b.Elapsed().Seconds()
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

	ctx := async.NoContext()
	msg := "hello, world"

	b.RunParallel(func(p *testing.PB) {
		req := testEchoRequest(b, msg)

		for p.Next() {
			result, st := client.Request(ctx, req)
			if !st.OK() {
				b.Fatal(st)
			}

			result.Release()
		}
	})

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

// Stream

func BenchmarkStream(b *testing.B) {
	streamMsg := []byte("hello, world")
	closeMsg := []byte("close")

	handle := func(ctx Context, ch ServerChannel) (ref.R[[]byte], status.Status) {
		for {
			msg, st := ch.Receive(ctx)
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

		return ref.NewFreer(buf.Bytes(), buf), status.OK
	}

	ctx := async.NoContext()
	server := testServer(b, handle)
	client := testClient(b, server)
	defer client.Close()

	b.ReportAllocs()
	b.ResetTimer()

	{
		req := testEchoRequest(b, "request")
		ch, st := client.Channel(ctx, req)
		if !st.OK() {
			b.Fatal(st)
		}
		defer ch.Free()

		for i := 0; i < b.N; i++ {
			st = ch.Send(ctx, streamMsg)
			if !st.OK() {
				b.Fatal(st)
			}
		}

		st = ch.Send(ctx, []byte(closeMsg))
		if !st.OK() {
			b.Fatal(st)
		}

		result, st := ch.Response(ctx)
		if !st.OK() {
			b.Fatal(st)
		}

		assert.Equal(b, "response", result.String().Unwrap())
	}

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

// RequestOneway

func BenchmarkRequestOneway(b *testing.B) {
	server := testEchoServer(b)
	client := testClient(b, server)
	defer client.Close()

	b.ReportAllocs()
	b.ResetTimer()

	ctx := async.NoContext()
	msg := "hello, world"
	req := testEchoRequest(b, msg)

	for i := 0; i < b.N; i++ {
		st := client.RequestOneway(ctx, req)
		if !st.OK() {
			b.Fatal(st)
		}
	}

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

func BenchmarkRequestOneway_Parallel(b *testing.B) {
	server := testEchoServer(b)
	client := testClient(b, server)
	defer client.Close()

	b.ReportAllocs()
	b.ResetTimer()

	ctx := async.NoContext()
	msg := "hello, world"

	b.RunParallel(func(p *testing.PB) {
		req := testEchoRequest(b, msg)

		for p.Next() {
			st := client.RequestOneway(ctx, req)
			if !st.OK() {
				b.Fatal(st)
			}
		}
	})

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}
