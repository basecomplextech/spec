package rpc

import (
	"testing"

	"github.com/basecomplextech/baselibrary/tests"
	"github.com/stretchr/testify/assert"
)

func testConnect(t tests.T, s *server) *conn {
	addr := s.Address()

	c, st := Connect(addr, s.logger)
	if !st.OK() {
		t.Fatal(st)
	}
	return c.(*conn)
}

// Request

func TestConn_Request__should_request_response(t *testing.T) {
	server := testEchoServer(t)

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
