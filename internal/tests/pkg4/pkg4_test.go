// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package pkg4

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/basecomplextech/spec/rpc"
	"github.com/stretchr/testify/assert"
)

func testServer(t tests.T, logger logging.Logger, service Service) rpc.Server {
	opts := rpc.Default()
	handler := NewServiceHandler(service)
	server := rpc.NewServer("localhost:0", handler, logger, opts)

	st := server.Start()
	if !st.OK() {
		t.Fatal(st)
	}

	cleanup := func() {
		select {
		case <-server.Stop():
		case <-time.After(time.Second):
			t.Fatal("server not stopped")
		}
	}
	t.Cleanup(cleanup)

	select {
	case <-server.Listening().Wait():
	case <-time.After(time.Second):
		t.Fatal("server not listening")
	}

	return server
}

func testClient(t tests.T, logger logging.Logger, server rpc.Server) ServiceClient {
	address := server.Address()
	client := rpc.NewClient(address, rpc.ClientMode_OnDemand, logger, server.Options())
	return NewServiceClient(client)
}

// Request

func TestService_Request(t *testing.T) {
	ctx := async.NoContext()
	logger := logging.TestLogger(t)
	service := newTestService()
	server := testServer(t, logger, service)
	client := testClient(t, logger, server)

	// method
	{
		st := client.Method(ctx)
		if !st.OK() {
			t.Fatal(st)
		}
	}

	// method1
	{
		w := NewServiceMethod1RequestWriter()
		w.Msg("hello")
		req, err := w.Build()
		if err != nil {
			t.Fatal(err)
		}

		st := client.Method1(ctx, req)
		if !st.OK() {
			t.Fatal(st)
		}
	}

	// method2
	{
		w := NewServiceMethod2RequestWriter()
		w.A(1)
		w.B(2)
		w.C(true)
		req, err := w.Build()
		if err != nil {
			t.Fatal(err)
		}

		resp_, st := client.Method2(ctx, req)
		if !st.OK() {
			t.Fatal(st)
		}
		resp := resp_.Unwrap()

		assert.Equal(t, int64(1), resp.A())
		assert.Equal(t, float64(2), resp.B())
		assert.True(t, resp.C())
	}

	// method3
	{
		w := NewRequestWriter()
		w.Msg("hello")

		req, err := w.Build()
		if err != nil {
			t.Fatal(err)
		}

		resp, st := client.Method3(ctx, req)
		if !st.OK() {
			t.Fatal(st)
		}
		resp.Release()
	}

	// method4
	{
		w := NewServiceMethod4RequestWriter()
		w.A00(true)
		w.A01(1)
		w.A10(10)
		w.A11(11)
		w.A20(1)

		req, err := w.Build()
		if err != nil {
			t.Fatal(err)
		}

		resp, st := client.Method4(ctx, req)
		if !st.OK() {
			t.Fatal(st)
		}

		assert.True(t, resp.Unwrap().Ok())
	}

	// method10
	{
		_, st := client.Method10(ctx)
		if !st.OK() {
			t.Fatal(st)
		}
	}

	// method11
	{
		resp, st := client.Method11(ctx)
		if !st.OK() {
			t.Fatal(st)
		}
		defer resp.Release()

		assert.Equal(t, "hello", resp.Unwrap().A50().Unwrap())
	}
}

// Channel

func TestService_Channel(t *testing.T) {
	ctx := async.NoContext()
	logger := logging.TestLogger(t)
	service := newTestService()
	server := testServer(t, logger, service)
	client := testClient(t, logger, server)

	// method20
	{
		w := NewServiceMethod20RequestWriter()
		w.A(1)
		w.B(2)
		w.C(true)

		req, err := w.Build()
		if err != nil {
			t.Fatal(err)
		}

		ch, st := client.Method20(ctx, req)
		if !st.OK() {
			t.Fatal(st)
		}
		defer ch.Free()

		resp, st := ch.Response(ctx)
		if !st.OK() {
			t.Fatal(st)
		}

		assert.Equal(t, int64(1), resp.A())
		assert.Equal(t, float64(2), resp.B())
		assert.True(t, resp.C())
	}

	// method21
	{
		w := NewRequestWriter()
		w.Msg("hello")
		req, err := w.Build()
		if err != nil {
			t.Fatal(err)
		}

		ch, st := client.Method21(ctx, req)
		if !st.OK() {
			t.Fatal(st)
		}
		defer ch.Free()

		w1 := NewInWriter()
		w1.A(1)
		w1.B(2)
		w1.C("3")
		msg, err := w1.Build()
		if err != nil {
			t.Fatal(err)
		}

		st = ch.Send(ctx, msg)
		if !st.OK() {
			t.Fatal(st)
		}

		resp, st := ch.Response(ctx)
		if !st.OK() {
			t.Fatal(st)
		}

		assert.Equal(t, "hello", resp.Msg().Unwrap())
	}

	// method22
	{
		w := NewRequestWriter()
		w.Msg("hello")
		req, err := w.Build()
		if err != nil {
			t.Fatal(err)
		}

		ch, st := client.Method22(ctx, req)
		if !st.OK() {
			t.Fatal(st)
		}
		defer ch.Free()

		msg, st := ch.Receive(ctx)
		if !st.OK() {
			t.Fatal(st)
		}
		assert.Equal(t, int64(1), msg.A())
		assert.Equal(t, float64(2), msg.B())
		assert.Equal(t, "3", msg.C().Unwrap())

		resp, st := ch.Response(ctx)
		if !st.OK() {
			t.Fatal(st)
		}

		assert.Equal(t, "hello", resp.Msg().Unwrap())
	}

	// method23
	{
		w := NewRequestWriter()
		w.Msg("hello")
		req, err := w.Build()
		if err != nil {
			t.Fatal(err)
		}

		ch, st := client.Method23(ctx, req)
		if !st.OK() {
			t.Fatal(st)
		}
		defer ch.Free()

		msg, st := ch.Receive(ctx)
		if !st.OK() {
			t.Fatal(st)
		}
		assert.Equal(t, int64(1), msg.A())
		assert.Equal(t, float64(2), msg.B())
		assert.Equal(t, "3", msg.C().Unwrap())

		w1 := NewInWriter()
		w1.A(1)
		w1.B(2)
		w1.C("3")
		msg1, err := w1.Build()
		if err != nil {
			t.Fatal(err)
		}

		st = ch.Send(ctx, msg1)
		if !st.OK() {
			t.Fatal(st)
		}

		resp, st := ch.Response(ctx)
		if !st.OK() {
			t.Fatal(st)
		}

		assert.Equal(t, "hello", resp.Msg().Unwrap())
	}
}

// Subservice

func TestService_Subservice(t *testing.T) {
	ctx := async.NoContext()
	logger := logging.TestLogger(t)
	service := newTestService()
	server := testServer(t, logger, service)
	client := testClient(t, logger, server)

	w0 := NewServiceSubserviceRequestWriter()
	w0.Id(bin.Int128(0, 123))
	req0, err := w0.Build()
	if err != nil {
		t.Fatal(err)
	}

	w1 := NewSubserviceHelloRequestWriter()
	w1.Msg("hello")
	req1, err := w1.Build()
	if err != nil {
		t.Fatal(err)
	}

	resp, st := client.Subservice(req0).Hello(ctx, req1)
	if !st.OK() {
		t.Fatal(st)
	}
	defer resp.Release()

	assert.Equal(t, "hello", resp.Unwrap().Msg().Unwrap())
}
