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
	return newClient(addr, s.logger, s.options)
}

// Flags

func TestClient_Connected_Disconnected__should_signal_on_state_changes(t *testing.T) {
	server := testRequestServer(t)
	ctx := async.NoContext()

	client := testClient(t, server)
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

	closed := client.connector.closed().Get()
	assert.True(t, closed)
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
		testChannelSend(t, ctx, ch, "hello, world")
		ch.Free()
	}

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

	assert.True(t, conn.Closed().Get())
}

// Conn

func TestClient_Conn__should_establish_connection(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	defer client.Close()

	ctx := async.NoContext()
	conn, st := client.Conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	conn.Free()
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
