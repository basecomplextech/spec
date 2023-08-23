package tcp

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/stretchr/testify/assert"
)

func TestTCP(t *testing.T) {
	handle := func(cancel <-chan struct{}, stream Stream) status.Status {
		for {
			msg, st := stream.Read(cancel)
			if !st.OK() {
				return st
			}
			if st := stream.Write(msg); !st.OK() {
				return st
			}
		}
	}

	logger := logging.TestLogger(t)
	server := newServerStreamHandler("localhost:0", StreamHandlerFunc(handle), logger)

	running, st := server.Run()
	if !st.OK() {
		t.Fatal(st)
	}
	defer async.CancelWait(running)

	select {
	case <-server.listening.Wait():
	case <-time.After(time.Second):
		t.Fatal("server not listening")
	}
	defer server.ln.Close()

	addr := server.listenAddress()
	conn, st := newClientCon(addr)
	if !st.OK() {
		t.Fatal(st)
	}

	stream, st := conn.Open(nil, nil)
	if !st.OK() {
		t.Fatal(st)
	}

	msg0 := []byte("hello, world")
	st = stream.Write(msg0)
	if !st.OK() {
		t.Fatal(st)
	}

	msg1, st := stream.Read(nil)
	if !st.OK() {
		t.Fatal(st)
	}

	assert.Equal(t, msg0, msg1)
}
