package tcp

import (
	"bytes"
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/status"
	"github.com/stretchr/testify/assert"
)

// Read

func TestStream_Read__should_read_message(t *testing.T) {
	server := testServer(t, func(s Stream) status.Status {
		st := s.Write(nil, []byte("hello, stream"))
		if !st.OK() {
			return st
		}
		return s.Close()
	})

	conn := testConnect(t, server)
	defer conn.Free()

	stream := testOpen(t, conn)
	defer stream.Free()

	st := stream.Write(nil, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	msg, st := stream.Read(nil)
	if !st.OK() {
		t.Fatal(st)
	}

	assert.Equal(t, []byte("hello, stream"), msg)
}

func TestStream_Read__should_return_end_when_stream_closed(t *testing.T) {
	server := testServer(t, func(s Stream) status.Status {
		st := s.Write(nil, []byte("hello, stream"))
		if !st.OK() {
			return st
		}
		return s.Close()
	})

	conn := testConnect(t, server)
	defer conn.Free()

	stream := testOpen(t, conn)
	stream.Close()

	_, st := stream.Read(nil)
	assert.Equal(t, status.End, st)
}

func TestStream_Read__should_read_pending_messages_even_when_closed(t *testing.T) {
	sent := make(chan struct{})
	server := testServer(t, func(s Stream) status.Status {
		defer close(sent)

		st := s.Write(nil, []byte("hello, stream"))
		if !st.OK() {
			return st
		}
		st = s.Write(nil, []byte("how are you?"))
		if !st.OK() {
			return st
		}
		return s.Close()
	})

	conn := testConnect(t, server)
	defer conn.Free()

	stream := testOpen(t, conn)
	defer stream.Free()

	st := stream.Write(nil, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	select {
	case <-sent:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	msg, st := stream.Read(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.Equal(t, []byte("hello, stream"), msg)

	msg, st = stream.Read(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.Equal(t, []byte("how are you?"), msg)

	_, st = stream.Read(nil)
	assert.Equal(t, status.End, st)
}

// Write

func TestStream_Write__should_send_message(t *testing.T) {
	var msg0 []byte
	done := make(chan struct{})

	server := testServer(t, func(s Stream) status.Status {
		defer close(done)

		msg, st := s.Read(nil)
		if !st.OK() {
			t.Fatal(st)
		}

		msg0 = bytes.Clone(msg)
		return s.Close()
	})

	conn := testConnect(t, server)
	defer conn.Free()

	stream := testOpen(t, conn)
	defer stream.Free()

	msg1 := []byte("hello, world")
	st := stream.Write(nil, msg1)
	if !st.OK() {
		t.Fatal(st)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	assert.Equal(t, msg1, msg0)
}

func TestStream_Write__should_send_open_message_if_not_sent(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	stream := testOpen(t, conn)
	defer stream.Free()

	assert.False(t, stream.openSent)

	st := stream.Write(nil, []byte("hello, world"))
	if !st.OK() {
		t.Fatal(st)
	}

	assert.True(t, stream.openSent)
}

func TestStream_Write__should_return_error_when_stream_closed(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	stream := testOpen(t, conn)
	defer stream.Free()
	stream.Close()

	st := stream.Write(nil, []byte("hello, world"))
	assert.Equal(t, statusStreamClosed, st)
}

// Close

func TestStream_Close__should_send_close_message(t *testing.T) {
	closed := make(chan struct{})
	server := testServer(t, func(s Stream) status.Status {
		defer close(closed)

		_, st := s.Read(nil)
		return st
	})

	conn := testConnect(t, server)
	defer conn.Free()

	stream := testOpen(t, conn)
	defer stream.Free()

	st := stream.Write(nil, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	st = stream.Close()
	if !st.OK() {
		t.Fatal(st)
	}

	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestStream_Close__should_close_incoming_queue(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	stream := testOpen(t, conn)
	defer stream.Free()

	st := stream.Close()
	if !st.OK() {
		t.Fatal(st)
	}

	assert.True(t, stream.reader.queue.Closed())
}

func TestStream_Close__should_ignore_when_already_closed(t *testing.T) {
	server := testServer(t, func(s Stream) status.Status {
		_, st := s.Read(nil)
		return st
	})

	conn := testConnect(t, server)
	defer conn.Free()

	stream := testOpen(t, conn)
	defer stream.Free()

	st := stream.Write(nil, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	st = stream.Close()
	if !st.OK() {
		t.Fatal(st)
	}

	st = stream.Close()
	if !st.OK() {
		t.Fatal(st)
	}
}
