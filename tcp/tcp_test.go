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
	opts := Default()
	logger := logging.TestLogger(t)
	server := newServer("localhost:0", handle, logger, opts)

	routine, st := server.Start()
	if !st.OK() {
		t.Fatal(st)
	}

	cleanup := func() {
		routine.Cancel()

		select {
		case <-routine.Wait():
		case <-time.After(time.Second):
			t.Fatal("stop timeout")
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
	handle := func(ch Channel) status.Status {
		msg, st := ch.ReadSync(nil)
		if !st.OK() {
			return st
		}
		if st := ch.Write(nil, msg); !st.OK() {
			return st
		}
		return ch.Close()
	}

	return testServer(t, handle)
}

func testConnect(t tests.T, s *server) *conn {
	addr := s.Address()

	c, st := Connect(addr, s.logger, s.options)
	if !st.OK() {
		t.Fatal(st)
	}
	return c.(*conn)
}

func testChannel(t tests.T, c Conn) *channel {
	ch, st := c.Channel(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	return ch.(*channel)
}

// Open/Close

func TestOpenClose(t *testing.T) {
	handle := func(ch Channel) status.Status {
		for {
			msg, st := ch.ReadSync(nil)
			if !st.OK() {
				return st
			}
			if st := ch.Write(nil, msg); !st.OK() {
				return st
			}
		}
	}

	server := testServer(t, handle)
	conn := testConnect(t, server)
	defer conn.Free()

	msg0 := []byte("hello, world")
	ch, st := conn.Channel(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	defer ch.Free()

	st = ch.Write(nil, msg0)
	if !st.OK() {
		t.Fatal(st)
	}

	msg1, st := ch.ReadSync(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	if st := ch.Close(); !st.OK() {
		t.Fatal(st)
	}

	assert.Equal(t, msg0, msg1)
}
