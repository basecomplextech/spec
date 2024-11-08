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
	"github.com/stretchr/testify/require"
)

func testClient(t tests.T, s *server) *client {
	return testClientMode(t, s, ClientMode_OnDemand)
}

func testClientMode(t tests.T, s *server, mode ClientMode) *client {
	addr := s.Address()
	c := newClient(addr, mode, s.logger, s.options)

	t.Cleanup(func() {
		c.Close()
	})
	return c
}

// NewClient

func TestNewClient__should_open_connection_when_autoconnect(t *testing.T) {
	server := testRequestServer(t)
	client := testClientMode(t, server, ClientMode_AutoConnect)

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
	opts := Default()
	opts.ClientMaxConns = 2
	opts.ClientConnChannels = 1

	server := testRequestServerOpts(t, opts)
	client := testClient(t, server)

	ctx := async.NoContext()
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
	assert.Same(t, cn0.unwrap().conn, cn1.unwrap().conn)
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

// Connect

func TestClient__should_retry_on_error_in_autoconnect_mode(t *testing.T) {
	server := testRequestServer(t)
	client := testClientMode(t, server, ClientMode_AutoConnect)
	server.Stop()

	ctx := async.NoContext()
	_, st := client.Conn(ctx)
	require.Equal(t, codeMpxError, st.Code)
	require.Contains(t, st.Message, "connection refused")

	time.Sleep(minConnectRetryTimeout)

	attempt := func() int {
		client.mu.Lock()
		defer client.mu.Unlock()

		return client.connectAttempt
	}()
	assert.True(t, attempt > 1)
}

func TestClient__should_reconnect_on_error_in_autoconnect_mode(t *testing.T) {
	server := testRequestServer(t)
	client := testClientMode(t, server, ClientMode_AutoConnect)
	ctx := async.NoContext()

	// Stop server
	server.Stop()
	select {
	case <-server.Stopped().Wait():
	case <-time.After(time.Second):
		t.Fatal("stop timeout")
	}

	// Try to connect
	_, st := client.Conn(ctx)
	require.Equal(t, codeMpxError, st.Code)
	require.Contains(t, st.Message, "connection refused")

	// Start server after some time
	time.Sleep(minConnectRetryTimeout)
	server.Start()
	select {
	case <-server.Listening().Wait():
	case <-time.After(100 * time.Second):
		t.Fatal("listen timeout")
	}

	// Connect again
	conn, st := client.Conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	conn.Free()
}
