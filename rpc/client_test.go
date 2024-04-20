package rpc

import (
	"bytes"
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/basecomplextech/spec"
	"github.com/stretchr/testify/assert"
)

func testClient(t tests.T, s *server) *client {
	addr := s.Address()
	return newClient(addr, s.logger, s.Server.Options())
}

// Close

func TestClient_Close__should_close_client(t *testing.T) {
	server := testEchoServer(t)
	client := testClient(t, server)

	st := client.Close()
	if !st.OK() {
		t.Fatal(st)
	}
}

// Request

func TestClient_Request__should_send_request_receive_response(t *testing.T) {
	server := testEchoServer(t)

	client := testClient(t, server)
	defer client.Close()

	ctx := async.NoContext()
	msg := "hello, world"
	req := testEchoRequest(t, msg)

	result, st := client.Request(ctx, req)
	if !st.OK() {
		t.Fatal(st)
	}
	defer result.Release()

	assert.Equal(t, msg, result.Unwrap().String().Unwrap())
}

// Write

func TestClient_Write__should_send_client_message_to_server(t *testing.T) {
	done := make(chan struct{})
	var message []byte

	handle := func(ctx async.Context, ch ServerChannel) (ref.R[[]byte], status.Status) {
		msg, st := ch.ReadSync(ctx)
		if !st.OK() {
			return nil, st
		}

		message = bytes.Clone(msg)
		close(done)
		return nil, status.OK
	}

	server := testServer(t, handle)
	client := testClient(t, server)
	defer client.Close()

	ctx := async.NoContext()
	req := testEchoRequest(t, "request")

	ch, st := client.Channel(ctx, req)
	if !st.OK() {
		t.Fatal(st)
	}

	st = ch.Write(ctx, []byte("hello, world"))
	if !st.OK() {
		t.Fatal(st)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	assert.Equal(t, []byte("hello, world"), message)
}

func TestClient_End__should_send_end_message_to_server(t *testing.T) {
	done := make(chan struct{})
	ended := false

	handle := func(ctx async.Context, ch ServerChannel) (ref.R[[]byte], status.Status) {
		msg, st := ch.ReadSync(ctx)
		if !st.OK() {
			return nil, st
		}
		assert.Equal(t, []byte("client message"), msg)

		_, st = ch.ReadSync(ctx)
		assert.Equal(t, status.CodeEnd, st.Code)

		close(done)
		ended = true
		return nil, status.OK
	}

	server := testServer(t, handle)
	client := testClient(t, server)
	defer client.Close()

	ctx := async.NoContext()
	req := testEchoRequest(t, "request")

	ch, st := client.Channel(ctx, req)
	if !st.OK() {
		t.Fatal(st)
	}

	st = ch.Write(ctx, []byte("client message"))
	if !st.OK() {
		t.Fatal(st)
	}

	st = ch.WriteEnd(ctx)
	if !st.OK() {
		t.Fatal(st)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	assert.True(t, ended)
}

// ReadSync

func TestClient_ReadSync__should_read_server_message(t *testing.T) {
	handle := func(ctx async.Context, ch ServerChannel) (ref.R[[]byte], status.Status) {
		st := ch.Write(ctx, []byte("hello, world"))
		if !st.OK() {
			return nil, st
		}
		return nil, status.OK
	}

	server := testServer(t, handle)
	client := testClient(t, server)
	defer client.Close()

	ctx := async.NoContext()
	req := testEchoRequest(t, "request")

	ch, st := client.Channel(ctx, req)
	if !st.OK() {
		t.Fatal(st)
	}

	msg, st := ch.ReadSync(ctx)
	if !st.OK() {
		t.Fatal(st)
	}

	assert.Equal(t, []byte("hello, world"), msg)
}

func TestClient_ReadSync__should_return_end_on_response(t *testing.T) {
	handle := func(ctx async.Context, ch ServerChannel) (ref.R[[]byte], status.Status) {
		st := ch.Write(ctx, []byte("server message"))
		if !st.OK() {
			return nil, st
		}

		buf := alloc.NewBuffer()
		w := spec.NewValueWriterBuffer(buf)
		w.String("response")
		if _, err := w.Build(); err != nil {
			return nil, status.WrapError(err)
		}
		return ref.NewFreer(buf.Bytes(), buf), status.OK
	}

	server := testServer(t, handle)
	client := testClient(t, server)
	defer client.Close()

	ctx := async.NoContext()
	req := testEchoRequest(t, "request")

	ch, st := client.Channel(ctx, req)
	if !st.OK() {
		t.Fatal(st)
	}

	msg, st := ch.ReadSync(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.Equal(t, []byte("server message"), msg)

	_, st = ch.ReadSync(ctx)
	assert.Equal(t, status.End, st)

	result, st := ch.Response(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	defer result.Release()

	assert.Equal(t, "response", result.Unwrap().String().Unwrap())
}

// Response

func TestClient_Response__should_receive_server_response(t *testing.T) {
	handle := func(ctx async.Context, ch ServerChannel) (ref.R[[]byte], status.Status) {
		buf := alloc.NewBuffer()
		w := spec.NewValueWriterBuffer(buf)
		w.String("hello, world")
		if _, err := w.Build(); err != nil {
			return nil, status.WrapError(err)
		}
		return ref.NewFreer(buf.Bytes(), buf), status.OK
	}

	server := testServer(t, handle)
	client := testClient(t, server)
	defer client.Close()

	ctx := async.NoContext()
	req := testEchoRequest(t, "request")

	ch, st := client.Channel(ctx, req)
	if !st.OK() {
		t.Fatal(st)
	}

	result, st := ch.Response(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	defer result.Release()

	assert.Equal(t, "hello, world", result.Unwrap().String().Unwrap())
}

func TestClient_Response__should_skip_message(t *testing.T) {
	handle := func(ctx async.Context, ch ServerChannel) (ref.R[[]byte], status.Status) {
		st := ch.Write(ctx, []byte("server message"))
		if !st.OK() {
			return nil, st
		}

		buf := alloc.NewBuffer()
		w := spec.NewValueWriterBuffer(buf)
		w.String("response")
		if _, err := w.Build(); err != nil {
			return nil, status.WrapError(err)
		}
		return ref.NewFreer(buf.Bytes(), buf), status.OK
	}

	server := testServer(t, handle)
	client := testClient(t, server)
	defer client.Close()

	ctx := async.NoContext()
	req := testEchoRequest(t, "request")

	ch, st := client.Channel(ctx, req)
	if !st.OK() {
		t.Fatal(st)
	}

	result, st := ch.Response(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	defer result.Release()

	assert.Equal(t, "response", result.Unwrap().String().Unwrap())
}

// Full

func TestClient_Channel__should_send_receive_messages_response(t *testing.T) {
	handle := func(ctx async.Context, ch ServerChannel) (ref.R[[]byte], status.Status) {
		st := ch.Write(ctx, []byte("server message"))
		if !st.OK() {
			return nil, st
		}

		msg, st := ch.ReadSync(ctx)
		if !st.OK() {
			return nil, st
		}
		assert.Equal(t, []byte("client message"), msg)

		_, st = ch.ReadSync(ctx)
		assert.Equal(t, status.End, st)

		st = ch.WriteEnd(ctx)
		if !st.OK() {
			return nil, st
		}

		buf := alloc.NewBuffer()
		w := spec.NewValueWriterBuffer(buf)
		w.String("response")
		if _, err := w.Build(); err != nil {
			return nil, status.WrapError(err)
		}
		return ref.NewFreer(buf.Bytes(), buf), status.OK
	}

	server := testServer(t, handle)
	client := testClient(t, server)
	defer client.Close()

	ctx := async.NoContext()
	req := testEchoRequest(t, "request")

	ch, st := client.Channel(ctx, req)
	if !st.OK() {
		t.Fatal(st)
	}

	msg, st := ch.ReadSync(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.Equal(t, []byte("server message"), msg)

	st = ch.Write(ctx, []byte("client message"))
	if !st.OK() {
		t.Fatal(st)
	}

	st = ch.WriteEnd(ctx)
	if !st.OK() {
		t.Fatal(st)
	}

	_, st = ch.ReadSync(ctx)
	assert.Equal(t, status.End, st)

	result, st := ch.Response(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	defer result.Release()

	assert.Equal(t, "response", result.Unwrap().String().Unwrap())
}
