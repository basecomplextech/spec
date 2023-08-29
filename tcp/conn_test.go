package tcp

import (
	"testing"

	"github.com/basecomplextech/baselibrary/status"
	"github.com/stretchr/testify/assert"
)

func TestConn_Open__should_open_new_stream(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	stream := testOpen(t, conn)
	defer stream.Free()

	st := stream.Write(nil, []byte("hello, world"))
	if !st.OK() {
		t.Fatal(st)
	}

	assert.True(t, stream.client)
	assert.False(t, stream.freed)
	assert.False(t, stream.closed)
	assert.True(t, stream.started)
	assert.True(t, stream.openSent)
}

func TestConn_Open__should_return_error_if_connection_is_closed(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	conn.Free()

	_, st := conn.Open(nil)
	assert.Equal(t, statusClosed, st)
}

// Free

func TestConn_Free__should_close_connection(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	conn.Free()

	assert.Equal(t, statusClosed, conn.st)
	assert.True(t, conn.writeQueue.Closed())
}

func TestConn_Free__should_close_streams(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	stream := testOpen(t, conn)
	conn.Free()

	_, st := stream.Read(nil)
	assert.Equal(t, status.End, st)
}

// HandleStream

func TestConn_handleStream__should_log_stream_panics(t *testing.T) {
	server := testServer(t, func(stream Stream) status.Status {
		panic("test")
	})
	conn := testConnect(t, server)
	defer conn.Free()

	stream := testOpen(t, conn)
	defer stream.Free()

	st := stream.Write(nil, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	_, st = stream.Read(nil)
	assert.Equal(t, status.End, st)
}

func TestConn_handleStream__should_log_stream_errors(t *testing.T) {
	server := testServer(t, func(stream Stream) status.Status {
		return status.Errorf("test stream error")
	})
	conn := testConnect(t, server)
	defer conn.Free()

	stream := testOpen(t, conn)
	defer stream.Free()

	st := stream.Write(nil, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	_, st = stream.Read(nil)
	assert.Equal(t, status.End, st)
}
