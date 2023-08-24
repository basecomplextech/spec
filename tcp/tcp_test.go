package tcp

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/stretchr/testify/assert"
)

func testServer(t *testing.T,
	address string,
	handler StreamHandlerFunc,
	logger logging.Logger,
) *server {
	return newServerStreamHandler(address, handler, logger)
}

func TestTCP(t *testing.T) {
	handle := func(cancel <-chan struct{}, stream Stream) status.Status {
		for {
			msg, st := stream.Read(cancel)
			if !st.OK() {
				return st
			}
			if st := stream.Write(cancel, msg); !st.OK() {
				return st
			}
		}
	}

	logger := logging.TestLogger(t)
	server := testServer(t, "localhost:0", handle, logger)

	run, st := server.Run()
	if !st.OK() {
		t.Fatal(st)
	}
	defer async.CancelWait(run)

	select {
	case <-server.listening.Wait():
	case <-time.After(time.Second):
		t.Fatal("server not listening")
	}

	addr := server.listenAddress()
	conn, st := Dial(addr, logger)
	if !st.OK() {
		t.Fatal(st)
	}
	defer conn.Free()

	stream, st := conn.Open(nil)
	if !st.OK() {
		t.Fatal(st)
	}

	msg0 := []byte("hello, world")
	st = stream.Write(nil, msg0)
	if !st.OK() {
		t.Fatal(st)
	}

	msg1, st := stream.Read(nil)
	if !st.OK() {
		t.Fatal(st)
	}

	assert.Equal(t, msg0, msg1)
}
