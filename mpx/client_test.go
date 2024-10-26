// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

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
	c := newClient(addr, s.logger, s.options)

	t.Cleanup(func() {
		c.Close()
	})
	return c
}

// NewClient

func TestNewClient__should_open_connection_when_autoconnect(t *testing.T) {
	server := testRequestServer(t)
	server.options.Client.AutoConnect = true
	client := testClient(t, server)

	select {
	case <-client.Connected().Wait():
	case <-time.After(time.Second):
		t.Fatal("connect timeout")
	}
}

// Flags

func TestClient__should_set_connected_flag_on_conn_opened(t *testing.T) {
	server := testRequestServer(t)
	client := testClient(t, server)
	ctx := async.NoContext()

	_, st := client.Conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}

	assert.True(t, client.connected_.IsSet())
	assert.False(t, client.disconnected_.IsSet())
}

func TestClient__should_set_disconnected_flag_when_all_conns_closed(t *testing.T) {
	server := testRequestServer(t)
	client := testClient(t, server)
	ctx := async.NoContext()

	conn, st := client.Conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	conn.Close()

	select {
	case <-client.disconnected_.Wait():
	case <-time.After(time.Second):
		t.Fatal("disconnect timeout")
	}

	assert.False(t, client.connected_.IsSet())
	assert.True(t, client.disconnected_.IsSet())
}

func TestClient__should_switch_flags_on_state_changes(t *testing.T) {
	server := testRequestServer(t)
	client := testClient(t, server)
	ctx := async.NoContext()

	_, st := client.Conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}

	select {
	case <-client.Connected().Wait():
	case <-time.After(time.Second):
		t.Fatal("connect timeout")
	}

	go func() {
		time.Sleep(time.Millisecond * 10)
		client.Close()
	}()

	select {
	case <-client.Disconnected().Wait():
	case <-time.After(time.Second):
		t.Fatal("disconnect timeout")
	}
}

// Close

func TestClient_Close__should_close_client(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	client.Close()

	closed := client.Closed().IsSet()
	assert.True(t, closed)
}

func TestClient_Close__should_close_connection(t *testing.T) {
	server := testRequestServer(t)
	client := testClient(t, server)
	ctx := async.NoContext()

	conn, st := client.Conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}

	client.Close()

	select {
	case <-conn.Closed().Wait():
	case <-time.After(time.Second):
		t.Fatal("close timeout")
	}

	assert.True(t, conn.Closed().IsSet())
}

// Conn

func TestClient_Conn__should_open_connection(t *testing.T) {
	server := testRequestServer(t)
	client := testClient(t, server)
	ctx := async.NoContext()

	conn, st := client.Conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	conn.Free()
}

func TestClient_Conn__should_return_existing_connection(t *testing.T) {
	server := testRequestServer(t)
	client := testClient(t, server)
	ctx := async.NoContext()

	conn, st := client.Conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}

	conn1, st := client.Conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.Same(t, conn, conn1)
}

func TestClient_Conn__should_reconnect_when_connection_closed(t *testing.T) {
	server := testRequestServer(t)
	client := testClient(t, server)
	ctx := async.NoContext()

	conn, st := client.Conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}

	conn.Close()
	select {
	case <-conn.Closed().Wait():
	case <-time.After(time.Second):
		t.Fatal("close timeout")
	}

	conn1, st := client.Conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.NotSame(t, conn, conn1)
}

func TestClient_Conn__should_open_more_connections_when_channels_target_reached(t *testing.T) {
	server := testRequestServer(t)
	ctx := async.NoContext()

	client := testClient(t, server)
	client.options.Client.MaxConns = 2
	client.options.Client.ConnChannels = 1

	conn, st := client.Conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	if _, st := conn.Channel(ctx); !st.OK() {
		t.Fatal(st)
	}
	if _, st := conn.Channel(ctx); !st.OK() {
		t.Fatal(st)
	}

	conn1, st := client.Conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.Same(t, conn, conn1)

	time.Sleep(50 * time.Millisecond)

	client.mu.Lock()
	n := len(client.conns)
	client.mu.Unlock()

	assert.Equal(t, 2, n)
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

	testChannelSend(t, ctx, ch, "hello, world")
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

	testChannelSend(t, ctx, ch0, "hello, world")

	ch1, st := client.Channel(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	defer ch1.Free()

	cn0 := ch0.(*channel)
	cn1 := ch1.(*channel)
	assert.Same(t, cn0.state.conn, cn1.state.conn)
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

		testChannelSend(t, ctx, ch, "hello, world")
		ch.Free()
	}

	{
		ctx := async.NoContext()
		conn, st := client.Conn(ctx)
		if !st.OK() {
			t.Fatal(st)
		}
		conn.Close()

		select {
		case <-conn.Closed().Wait():
		case <-time.After(time.Second):
			t.Fatal("close timeout")
		}
	}

	ctx := async.NoContext()
	ch, st := client.Channel(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	defer ch.Free()

	testChannelSend(t, ctx, ch, "hello, world")
}
