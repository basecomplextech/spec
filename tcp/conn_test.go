package tcp

import (
	"testing"

	"github.com/basecomplextech/baselibrary/status"
	"github.com/stretchr/testify/assert"
)

func TestConn_Open__should_open_new_channel(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	st := ch.Write(nil, []byte("hello, world"))
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

	_, st := conn.Channel(nil)
	assert.Equal(t, statusConnClosed, st)
}

// Free

func TestConn_Free__should_close_connection(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	conn.Free()

	assert.Equal(t, statusConnClosed, conn.socket.st)
	assert.True(t, conn.writeQueue.Closed())
}

func TestConn_Free__should_close_chs(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	conn.Free()

	_, st := ch.Read(nil)
	assert.Equal(t, status.End, st)
}

// HandleChannel

func TestConn_handleChannel__should_log_ch_panics(t *testing.T) {
	server := testServer(t, func(ch Channel) status.Status {
		panic("test")
	})
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	st := ch.Write(nil, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	_, st = ch.Read(nil)
	assert.Equal(t, status.End, st)
}

func TestConn_handleChannel__should_log_ch_errors(t *testing.T) {
	server := testServer(t, func(ch Channel) status.Status {
		return status.Errorf("test ch error")
	})
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	st := ch.Write(nil, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	_, st = ch.Read(nil)
	assert.Equal(t, status.End, st)
}
