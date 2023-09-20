package pkg4

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/basecomplextech/spec/rpc"
	"github.com/stretchr/testify/assert"
)

func testServer(t tests.T, logger logging.Logger, service Service) rpc.Server {
	handler := NewServiceHandler(service)
	server := rpc.NewServer("localhost:0", handler, logger)

	routine, st := server.Start()
	if !st.OK() {
		t.Fatal(st)
	}

	cleanup := func() {
		routine.Cancel()

		select {
		case <-routine.Wait():
		case <-time.After(time.Second):
			t.Fatal("server not stopped")
		}
	}
	t.Cleanup(cleanup)

	select {
	case <-server.Listening():
	case <-time.After(time.Second):
		t.Fatal("server not listening")
	}

	return server
}

func testClient(t tests.T, logger logging.Logger, server rpc.Server) *ServiceClient {
	address := server.Address()
	client := rpc.NewClient(address, logger)
	return NewServiceClient(client)
}

// Request

func TestService_Request(t *testing.T) {
	logger := logging.TestLogger(t)
	service := newTestService()
	server := testServer(t, logger, service)
	client := testClient(t, logger, server)

	// method
	{
		st := client.Method(nil)
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

		st := client.Method1(nil, req)
		if !st.OK() {
			t.Fatal(st)
		}
	}

	// method2
	{
		a, b, c, st := client.Method2(nil, 1, 2, true)
		if !st.OK() {
			t.Fatal(st)
		}

		assert.Equal(t, int64(1), a)
		assert.Equal(t, float64(2), b)
		assert.True(t, c)
	}

	// method3
	{
		w := NewServiceMethod3RequestWriter()
		w.A00(true)
		w.A01(1)
		w.A10(10)
		w.A11(11)
		w.A20(1)

		req, err := w.Build()
		if err != nil {
			t.Fatal(err)
		}

		ok, st := client.Method3(nil, req)
		if !st.OK() {
			t.Fatal(st)
		}

		assert.True(t, ok)
	}

	// method10
	{
		_, _, _, _, _, _, _, _, _, _, _, _, _, st := client.Method10(nil)
		if !st.OK() {
			t.Fatal(st)
		}
	}

	// method11
	{
		resp, st := client.Method11(nil)
		if !st.OK() {
			t.Fatal(st)
		}
		defer resp.Release()

		assert.Equal(t, "hello", resp.Unwrap().A50().Unwrap())
	}
}

// Channel

func TestService_Channel(t *testing.T) {
	logger := logging.TestLogger(t)
	service := newTestService()
	server := testServer(t, logger, service)
	client := testClient(t, logger, server)

	// method20
	{
		ch, st := client.Method20(nil, 1, 2, true)
		if !st.OK() {
			t.Fatal(st)
		}
		defer ch.Free()

		a, b, c, st := ch.Response(nil)
		if !st.OK() {
			t.Fatal(st)
		}

		assert.Equal(t, int64(1), a)
		assert.Equal(t, float64(2), b)
		assert.True(t, c)
	}

	// method21
	{
		w := NewRequestWriter()
		w.Msg("hello")
		req, err := w.Build()
		if err != nil {
			t.Fatal(err)
		}

		ch, st := client.Method21(nil, req)
		if !st.OK() {
			t.Fatal(st)
		}
		defer ch.Free()

		msg, st := ch.Receive(nil)
		if !st.OK() {
			t.Fatal(st)
		}
		assert.Equal(t, int64(1), msg.A())
		assert.Equal(t, float64(2), msg.B())
		assert.Equal(t, "3", msg.C().Unwrap())

		resp, st := ch.Response(nil)
		if !st.OK() {
			t.Fatal(st)
		}
		defer resp.Release()

		assert.Equal(t, "hello", resp.Unwrap().Msg().Unwrap())
	}

	// method22
	{
		w := NewRequestWriter()
		w.Msg("hello")
		req, err := w.Build()
		if err != nil {
			t.Fatal(err)
		}

		ch, st := client.Method22(nil, req)
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

		st = ch.Send(nil, msg)
		if !st.OK() {
			t.Fatal(st)
		}

		resp, st := ch.Response(nil)
		if !st.OK() {
			t.Fatal(st)
		}
		defer resp.Release()

		assert.Equal(t, "hello", resp.Unwrap().Msg().Unwrap())
	}

	// method23
	{
		w := NewRequestWriter()
		w.Msg("hello")
		req, err := w.Build()
		if err != nil {
			t.Fatal(err)
		}

		ch, st := client.Method23(nil, req)
		if !st.OK() {
			t.Fatal(st)
		}
		defer ch.Free()

		msg, st := ch.Receive(nil)
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

		st = ch.Send(nil, msg1)
		if !st.OK() {
			t.Fatal(st)
		}

		resp, st := ch.Response(nil)
		if !st.OK() {
			t.Fatal(st)
		}
		defer resp.Release()

		assert.Equal(t, "hello", resp.Unwrap().Msg().Unwrap())
	}
}

// Subservice

func TestService_Subservice(t *testing.T) {
	logger := logging.TestLogger(t)
	service := newTestService()
	server := testServer(t, logger, service)
	client := testClient(t, logger, server)

	sub, st := client.Subservice(bin.Bin128FromInt(123))
	if !st.OK() {
		t.Fatal(st)
	}

	w := NewSubserviceHelloRequestWriter()
	w.Msg("hello")
	req, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}

	resp, st := sub.Hello(nil, req)
	if !st.OK() {
		t.Fatal(st)
	}
	defer resp.Release()

	assert.Equal(t, "hello", resp.Unwrap().Msg().Unwrap())
}
