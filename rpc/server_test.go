package rpc

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/prpc"
	"github.com/stretchr/testify/assert"
)

func testServer(t tests.T, handle HandleFunc) *server {
	logger := logging.TestLogger(t)
	server := newServer("localhost:0", handle, logger)

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

func testEchoServer(t tests.T) *server {
	handle := func(cancel <-chan struct{}, ch ServerChannel, req prpc.Request) (*alloc.Buffer, status.Status) {
		call := req.Calls().Get(0)
		msg := call.Args().String().Unwrap()

		buf := alloc.NewBuffer()
		w := spec.NewValueWriterBuffer(buf)
		w.String(msg)
		if _, err := w.Build(); err != nil {
			return nil, status.WrapError(err)
		}
		return buf, status.OK
	}

	return testServer(t, handle)
}

func testEchoRequest(t tests.T, msg string) prpc.Request {
	w := prpc.NewRequestWriter()
	calls := w.Calls()
	{
		call := calls.Add()
		call.Method("echo")
		call.Args().String(msg)
		if err := call.End(); err != nil {
			t.Fatal(err)
		}
	}
	if err := calls.End(); err != nil {
		t.Fatal(err)
	}

	preq, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	return preq
}

// handleRequest

func TestServer_handleRequest__should_handle_panics_and_send_error_response(t *testing.T) {
	handle := func(cancel <-chan struct{}, ch ServerChannel, req prpc.Request) (*alloc.Buffer, status.Status) {
		panic("test panic")
	}

	server := testServer(t, handle)
	client := testClient(t, server)
	defer client.Close()

	req := testEchoRequest(t, "request")
	_, st := client.Request(nil, req)

	assert.Equal(t, status.CodeError, st.Code)
	assert.Equal(t, "test panic", st.Message)
}

func TestServer_handleRequest__should_handle_errors(t *testing.T) {
	handle := func(cancel <-chan struct{}, ch ServerChannel, req prpc.Request) (*alloc.Buffer, status.Status) {
		return nil, status.Unauthorized("test unauthorized")
	}

	server := testServer(t, handle)
	client := testClient(t, server)
	defer client.Close()

	req := testEchoRequest(t, "request")
	_, st := client.Request(nil, req)

	assert.Equal(t, status.CodeUnauthorized, st.Code)
	assert.Equal(t, "test unauthorized", st.Message)
}
