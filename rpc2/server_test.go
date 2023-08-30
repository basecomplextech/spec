package rpc

import (
	"time"

	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/basecomplextech/spec/proto/prpc"
)

func testServer(t tests.T, handle HandleFunc) *server {
	logger := logging.TestLogger(t)
	server := newServer("localhost:0", handle, logger)

	run, st := server.Run()
	if !st.OK() {
		t.Fatal(st)
	}

	cleanup := func() {
		run.Cancel()

		select {
		case <-run.Wait():
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
	return testServer(t, func(cancel <-chan struct{}, req prpc.Request) (*Response, status.Status) {
		call := req.Calls().Get(0)
		arg := call.Args().Get(0)
		name := arg.Name().Unwrap()
		value := arg.Value().String().Unwrap()

		resp := NewResponse()
		result := resp.Add()
		result.Name(name)
		result.Value().String(value)
		if err := result.End(); err != nil {
			return nil, status.WrapError(err)
		}
		return resp, status.OK
	})
}

func testEchoRequest(t tests.T, msg string) *Request {
	req := NewRequest()
	call := req.Call("echo")
	args := call.Args()

	arg := args.Add()
	arg.Name("msg")
	arg.Value().String("hello, world")
	if err := arg.End(); err != nil {
		t.Fatal(err)
	}

	if err := args.End(); err != nil {
		t.Fatal(err)
	}
	if err := call.End(); err != nil {
		t.Fatal(err)
	}
	return req
}
