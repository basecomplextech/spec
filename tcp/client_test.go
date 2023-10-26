package tcp

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/tests"
	"github.com/stretchr/testify/assert"
)

func testClient(t tests.T, s *server) *client {
	addr := s.Address()
	return newClient(addr, s.logger, s.options)
}

// Connect

func TestClient_Connect_should_connect_to_server(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	future := client.Connect(nil)

	select {
	case <-future.Wait():
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	st := future.Status()
	if !st.OK() {
		t.Fatal(st)
	}

	assert.NotNil(t, client.conn)
	assert.Nil(t, client.connecting)
	assert.True(t, client.connected.IsSet())
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
		ch, st := client.Channel(nil)
		if !st.OK() {
			t.Fatal(st)
		}
		ch.Free()
	}

	conn := client.conn
	client.Close()

	assert.Equal(t, statusConnClosed, conn.socket.st)
}

// Channel

func TestClient_Channel__should_connect_to_server(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	defer client.Close()

	ch, st := client.Channel(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	defer ch.Free()
}

func TestClient_Channel__should_reuse_open_connection(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	defer client.Close()

	ch0, st := client.Channel(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	defer ch0.Free()

	ch1, st := client.Channel(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	defer ch1.Free()

	cn0 := ch0.(*channel)
	cn1 := ch1.(*channel)
	assert.Same(t, cn0.conn, cn1.conn)
}

func TestClient_Channel__should_return_error_if_client_is_closed(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	client.Close()

	_, st := client.Channel(nil)
	assert.Equal(t, statusClientClosed, st)
}

func TestClient_Channel__should_reconnect_if_connection_closed(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	defer client.Close()

	{
		conn, st := client.Channel(nil)
		if !st.OK() {
			t.Fatal(st)
		}
		conn.Free()
	}

	{
		conn := client.conn
		conn.close()
	}

	conn, st := client.Channel(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	defer conn.Free()
}
