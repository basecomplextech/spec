// Copyright 2023 Ivan Korobkov. All rights reserved.

package mpx

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/stretchr/testify/assert"
)

func TestConn_Open__should_open_new_channel(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	ctx := async.NoContext()
	st := ch.Send(ctx, []byte("hello, world"))
	if !st.OK() {
		t.Fatal(st)
	}

	assert.True(t, ch.state.client)
	assert.True(t, ch.state.sendOpen)
	assert.False(t, ch.state.sendClose)
}

func TestConn_Open__should_return_error_if_connection_is_closed(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	conn.Free()

	ctx := async.NoContext()
	_, st := conn.Channel(ctx)
	assert.Equal(t, statusConnClosed, st)
}

// Free

func TestConn_Free__should_close_connection(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	conn.Free()

	select {
	case <-conn.closed.Wait():
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	assert.True(t, conn.closed.Get())
	assert.True(t, conn.writeq.Closed())
}

func TestConn_Free__should_close_channels(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	conn.Free()

	ctx := async.NoContext()
	_, st := ch.Receive(ctx)
	assert.Equal(t, status.End, st)
}

func TestConn_Free__should_notify_listeners(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	var notified bool
	unsub := conn.OnClosed(func() {
		notified = true
	})
	defer unsub()

	conn.Free()
	select {
	case <-conn.closed.Wait():
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	assert.True(t, notified)
}

// HandleChannel

func TestConn_handleChannel__should_log_channel_panics(t *testing.T) {
	server := testServer(t, func(ctx Context, ch Channel) status.Status {
		panic("test")
	})
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	ctx := async.NoContext()
	st := ch.Send(ctx, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	_, st = ch.Receive(ctx)
	assert.Equal(t, status.End, st)
}

func TestConn_handleChannel__should_log_channel_errors(t *testing.T) {
	server := testServer(t, func(ctx Context, ch Channel) status.Status {
		return status.Errorf("test ch error")
	})
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	ctx := async.NoContext()
	st := ch.Send(ctx, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	_, st = ch.Receive(ctx)
	assert.Equal(t, status.End, st)
}
