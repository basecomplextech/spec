package tcp

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
	server := testRequestServer(t)

	client := testClient(t, server)
	client.Close()

	assert.True(t, client.closed)
}

func TestClient_Close__should_close_connection(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	defer client.Close()

	{
		conn, st := client.Connect(nil)
		if !st.OK() {
			t.Fatal(st)
		}
		conn.Free()
	}

	conn := client.conn.Unwrap()
	client.Close()

	assert.Equal(t, statusClosed, conn.st)
}

// Connect

func TestClient_Connect__should_connect_to_server(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	defer client.Close()

	conn, st := client.Connect(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	defer conn.Free()
}

func TestClient_Connect__should_reuse_open_connection(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	defer client.Close()

	conn1, st := client.Connect(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	defer conn1.Free()

	conn2, st := client.Connect(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	defer conn2.Free()

	cn0 := conn1.(*clientConn)
	cn1 := conn2.(*clientConn)
	assert.Same(t, cn0.conn, cn1.conn)
}

func TestClient_Connect__should_return_error_if_client_is_closed(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	client.Close()

	_, st := client.Connect(nil)
	assert.Equal(t, statusClientClosed, st)
}

func TestClient_Connect__should_reconnect_if_connection_closed(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	defer client.Close()

	{
		conn, st := client.Connect(nil)
		if !st.OK() {
			t.Fatal(st)
		}
		conn.Free()
	}

	{
		conn := client.conn.Unwrap()
		conn.close()
	}

	conn, st := client.Connect(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	defer conn.Free()
}
