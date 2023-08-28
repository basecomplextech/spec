package tcp

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/stretchr/testify/assert"
)

func testServer(t tests.T, handle HandlerFunc) (*server, func()) {
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

	select {
	case <-server.listening.Wait():
	case <-time.After(time.Second):
		t.Fatal("server not listening")
	}

	return server, cleanup
}

func testConnect(t tests.T, s *server) *conn {
	addr := s.listenAddress()

	c, st := Connect(addr, s.logger)
	if !st.OK() {
		t.Fatal(st)
	}
	return c.(*conn)
}

// Open/Close

func TestOpenClose(t *testing.T) {
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

	server, cleanup := testServer(t, handle)
	defer cleanup()

	conn := testConnect(t, server)
	defer conn.Free()

	msg0 := []byte("hello, world")
	stream, st := conn.Open(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	defer stream.Free()

	st = stream.Write(nil, msg0)
	if !st.OK() {
		t.Fatal(st)
	}

	msg1, st := stream.Read(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	if st := stream.Close(nil); !st.OK() {
		t.Fatal(st)
	}

	assert.Equal(t, msg0, msg1)
}
