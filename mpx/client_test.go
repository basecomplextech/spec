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

// Connect

func TestClient_Connect__should_start_connector(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	client.Connect()

	select {
	case <-client.Connected().Wait():
	case <-time.After(time.Second):
		t.Fatal("connect timeout")
	}

	assert.NotNil(t, client.conn)
	assert.NotNil(t, client.connector)
	assert.True(t, client.connected_.Get())
	assert.False(t, client.disconnected_.Get())
}

// Flags

func TestClient_Connected_Disconnected__should_signal_on_state_changes(t *testing.T) {
	server := testRequestServer(t)

	client := testClient(t, server)
	client.Connect()

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
		testChannelSend(t, ctx, ch, "hello, world")
		ch.Free()
	}

	conn := client.conn
	client.Close()

	select {
	case <-conn.closed.Wait():
	case <-time.After(time.Second):
		t.Fatal("close timeout")
	}

	assert.True(t, conn.closed.Get())
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
		conn := client.conn
		conn.Close()
	}

	ctx := async.NoContext()
	ch, st := client.Channel(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	defer ch.Free()

	testChannelSend(t, ctx, ch, "hello, world")
}

func TestClient_Channel__should_await_connection(t *testing.T) {
	server := testRequestServer(t)
	select {
	case <-server.Stop():
	case <-time.After(time.Second):
		t.Fatal("stop timeout")
	}

	client := testClient(t, server)
	defer client.Close()

	go func() {
		time.Sleep(time.Millisecond * 100)
		server.Start()

		select {
		case <-server.Listening().Wait():
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

	testChannelSend(t, ctx, ch, "hello, world")
}
