package tcp

import (
	"testing"

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
	st := ch.Write(ctx, []byte("hello, world"))
	if !st.OK() {
		t.Fatal(st)
	}

	assert.True(t, ch.client)
	assert.False(t, ch.state.closed)
	assert.True(t, ch.state.started)
	assert.True(t, ch.state.newSent)
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

	assert.Equal(t, statusConnClosed, conn.socket.st)
	assert.True(t, conn.writeq.Closed())
}

func TestConn_Free__should_close_channels(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	conn.Free()

	ctx := async.NoContext()
	_, st := ch.ReadSync(ctx)
	assert.Equal(t, status.End, st)
}

// HandleChannel

func TestConn_handleChannel__should_log_channel_panics(t *testing.T) {
	server := testServer(t, func(ctx async.Context, ch Channel) status.Status {
		panic("test")
	})
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	ctx := async.NoContext()
	st := ch.Write(ctx, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	_, st = ch.ReadSync(ctx)
	assert.Equal(t, status.End, st)
}

func TestConn_handleChannel__should_log_channel_errors(t *testing.T) {
	server := testServer(t, func(ctx async.Context, ch Channel) status.Status {
		return status.Errorf("test ch error")
	})
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	ctx := async.NoContext()
	st := ch.Write(ctx, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	_, st = ch.ReadSync(ctx)
	assert.Equal(t, status.End, st)
}
