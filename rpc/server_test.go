package rpc

import (
	"time"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/basecomplextech/spec/proto/prpc"
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
		msg := string(call.Args().Unwrap())

		buf := alloc.NewBuffer()
		buf.WriteString(msg)
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
		call.Args([]byte(msg))
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
