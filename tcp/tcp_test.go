package tcp

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/baselibrary/tests"
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
	case <-server.listening.Wait():
	case <-time.After(time.Second):
		t.Fatal("server not listening")
	}

	return server
}

func testRequestServer(t tests.T) *server {
	handle := func(stream Stream) status.Status {
		msg, st := stream.Read(nil)
		if !st.OK() {
			return st
		}
		if st := stream.Write(nil, msg); !st.OK() {
			return st
		}
		return stream.Close()
	}

	return testServer(t, handle)
}

func testConnect(t tests.T, s *server) *conn {
	addr := s.Address()

	c, st := Connect(addr, s.logger)
	if !st.OK() {
		t.Fatal(st)
	}
	return c.(*conn)
}

func testStream(t tests.T, c Conn) *stream {
	s, st := c.Stream(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	return s.(*stream)
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

	server := testServer(t, handle)
	conn := testConnect(t, server)
	defer conn.Free()

	msg0 := []byte("hello, world")
	stream, st := conn.Stream(nil)
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
	if st := stream.Close(); !st.OK() {
		t.Fatal(st)
	}

	assert.Equal(t, msg0, msg1)
}
