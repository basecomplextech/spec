package tcp

import (
	"testing"

	"github.com/basecomplextech/baselibrary/status"
	"github.com/stretchr/testify/assert"
)

const address = "localhost:9999"

func TestTCP(t *testing.T) {
	serve := func(s Stream) status.Status {
		for {
			msg, st := s.Read(nil)
			if !st.OK() {
				return st
			}
			if st := s.Write(msg); !st.OK() {
				return st
			}
		}
	}

	server := newServer(address, HandlerFunc(serve))
	go server.run()

	<-server.listening.Wait()
	defer server.ln.Close()

	conn, st := newClientCon(address)
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
