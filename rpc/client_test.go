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

// Connect

func TestClient_Connect__should_return_connection(t *testing.T) {
	server := testEchoServer(t)

	client := testClient(t, server)
	defer client.Close()

	conn := testConnect(t, server)
	defer conn.Free()

	msg := "hello, world"
	req := testEchoRequest(t, msg)
	defer req.Free()

	resp, st := conn.Request(nil, req)
	if !st.OK() {
		t.Fatal(st)
	}
	defer resp.Free()

	result := resp.Unwrap().Results().Get(0)
	msg1 := result.Value().String().Unwrap()
	assert.Equal(t, msg, msg1)
}
