package mpx

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/stretchr/testify/assert"
)

func testClient(t tests.T, s *server) *client {
	addr := s.Address()
	return newClient(addr, s.logger, s.options)
}

// Connect

func TestClient_Connect__should_start_connector(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	client.Connect()

	select {
	case <-client.Connected():
	case <-time.After(time.Second):
		t.Fatal("connect timeout")
	}

	assert.NotNil(t, client.conn)
	assert.NotNil(t, client.connector)
	assert.True(t, client.connected_.IsSet())
	assert.False(t, client.disconnected_.IsSet())
}

// Flags

func TestClient_Connected_Disconnected__should_signal_on_state_changes(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	client.Connect()

	select {
	case <-client.Connected():
	case <-time.After(time.Second):
		t.Fatal("connect timeout")
	}

	go func() {
		time.Sleep(time.Millisecond * 10)
		client.Close()
	}()

	select {
	case <-client.Disconnected():
	case <-time.After(time.Second):
		t.Fatal("disconnect timeout")
	}
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
		ctx := async.NoContext()
		ch, st := client.Channel(ctx)
		if !st.OK() {
			t.Fatal(st)
		}
		testChannelWrite(t, ctx, ch, "hello, world")
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

	ctx := async.NoContext()
	ch, st := client.Channel(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	defer ch.Free()

	testChannelWrite(t, ctx, ch, "hello, world")
}

func TestClient_Channel__should_reuse_open_connection(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	defer client.Close()

	ctx := async.NoContext()
	ch0, st := client.Channel(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	defer ch0.Free()

	testChannelWrite(t, ctx, ch0, "hello, world")

	ch1, st := client.Channel(ctx)
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

	ctx := async.NoContext()
	_, st := client.Channel(ctx)
	assert.Equal(t, statusClientClosed, st)
}

func TestClient_Channel__should_reconnect_if_connection_closed(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	defer client.Close()

	{
		ctx := async.NoContext()
		ch, st := client.Channel(ctx)
		if !st.OK() {
			t.Fatal(st)
		}

		testChannelWrite(t, ctx, ch, "hello, world")
		ch.Free()
	}

	{
		conn := client.conn
		conn.Close()
	}

	ctx := async.NoContext()
	ch, st := client.Channel(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	defer ch.Free()

	testChannelWrite(t, ctx, ch, "hello, world")
}

func TestClient_Channel__should_await_connection(t *testing.T) {
	server := testRequestServer(t)
	select {
	case <-server.Stop().Wait():
	case <-time.After(time.Second):
		t.Fatal("stop timeout")
	}

	client := testClient(t, server)
	defer client.Close()

	go func() {
		time.Sleep(time.Millisecond * 100)
		server.Start()

		select {
		case <-server.Listening():
		case <-time.After(time.Second):
			t.Fatal("start timeout")
		}

		client.address = server.Address()
	}()

	ctx := async.NoContext()
	ch, st := client.Channel(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	defer ch.Free()

	testChannelWrite(t, ctx, ch, "hello, world")
}
