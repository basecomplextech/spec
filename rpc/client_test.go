package rpc

import (
	"testing"

	"github.com/basecomplextech/baselibrary/tests"
	"github.com/stretchr/testify/assert"
)

func testClient(t tests.T, s *server) *client {
	addr := s.Address()
	return newClient(addr, s.logger)
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

	msg := "hello, world"
	req := testEchoRequest(t, msg)

	buf, st := client.Request(nil, req)
	if !st.OK() {
		t.Fatal(st)
	}
	defer buf.Free()

	msg1 := string(buf.Bytes())
	assert.Equal(t, msg, msg1)
}
