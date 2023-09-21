package tcp

import (
	"bytes"
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/status"
	"github.com/stretchr/testify/assert"
)

// Read

func TestChannel_Read__should_read_message(t *testing.T) {
	server := testServer(t, func(s Channel) status.Status {
		st := s.Write(nil, []byte("hello, channel"))
		if !st.OK() {
			return st
		}
		return s.Close()
	})

	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	st := ch.Write(nil, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	msg, st := ch.Read(nil)
	if !st.OK() {
		t.Fatal(st)
	}

	assert.Equal(t, []byte("hello, channel"), msg)
}

func TestChannel_Read__should_return_end_when_channel_closed(t *testing.T) {
	server := testServer(t, func(s Channel) status.Status {
		return s.Close()
	})

	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	ch.Close()

	_, st := ch.Read(nil)
	assert.Equal(t, status.End, st)
}

func TestChannel_Read__should_read_pending_messages_even_when_closed(t *testing.T) {
	sent := make(chan struct{})
	server := testServer(t, func(s Channel) status.Status {
		defer close(sent)

		st := s.Write(nil, []byte("hello, channel"))
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

	ch := testChannel(t, conn)
	defer ch.Free()

	st := ch.Write(nil, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	select {
	case <-sent:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	msg, st := ch.Read(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.Equal(t, []byte("hello, channel"), msg)

	msg, st = ch.Read(nil)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.Equal(t, []byte("how are you?"), msg)

	_, st = ch.Read(nil)
	assert.Equal(t, status.End, st)
}

func TestChannel_Read__should_increment_read_bytes(t *testing.T) {
	server := testServer(t, func(s Channel) status.Status {
		st := s.Write(nil, []byte("hello, channel"))
		if !st.OK() {
			return st
		}
		return s.Close()
	})

	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	st := ch.Write(nil, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}
	assert.Equal(t, 0, ch.state.readBytes)

	_, st = ch.Read(nil)
	if !st.OK() {
		t.Fatal(st)
	}

	assert.Equal(t, 53, ch.state.readBytes)
}

// Write

func TestChannel_Write__should_send_message(t *testing.T) {
	var msg0 []byte
	done := make(chan struct{})

	server := testServer(t, func(s Channel) status.Status {
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

	ch := testChannel(t, conn)
	defer ch.Free()

	msg1 := []byte("hello, world")
	st := ch.Write(nil, msg1)
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

func TestChannel_Write__should_send_new_message_if_not_sent(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	assert.False(t, ch.state.newSent)

	st := ch.Write(nil, []byte("hello, world"))
	if !st.OK() {
		t.Fatal(st)
	}

	assert.True(t, ch.state.newSent)
}

func TestChannel_Write__should_return_error_when_ch_closed(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()
	ch.Close()

	st := ch.Write(nil, []byte("hello, world"))
	assert.Equal(t, statusChannelClosed, st)
}

func TestChannel_Write__should_decrement_write_window(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()
	assert.Equal(t, ch.state.window, ch.state.writeWindow)

	st := ch.Write(nil, []byte("hello, world"))
	if !st.OK() {
		t.Fatal(st)
	}

	assert.Equal(t, ch.state.window-51, ch.state.writeWindow)
}

func TestChannel_Write__should_block_when_write_window_not_enough(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	assert.Equal(t, ch.state.window, ch.state.writeWindow)
	msg := bytes.Repeat([]byte("a"), (ch.state.window / 3))

	st := ch.Write(nil, msg)
	if !st.OK() {
		t.Fatal(st)
	}
	st = ch.Write(nil, msg)
	if !st.OK() {
		t.Fatal(st)
	}

	timeout := make(chan struct{})
	go func() {
		defer close(timeout)
		time.Sleep(time.Millisecond * 100)
	}()

	st = ch.Write(timeout, msg)
	assert.Equal(t, status.Cancelled, st)
}

func TestChannel_Write__should_wait_write_window_increment(t *testing.T) {
	server := testServer(t, func(ch Channel) status.Status {
		if _, st := ch.Read(nil); !st.OK() {
			return st
		}
		if _, st := ch.Read(nil); !st.OK() {
			return st
		}
		return status.OK
	})
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	assert.Equal(t, ch.state.window, ch.state.writeWindow)
	msg := bytes.Repeat([]byte("a"), (ch.state.window/2)+1)

	st := ch.Write(nil, msg)
	if !st.OK() {
		t.Fatal(st)
	}

	st = ch.Write(nil, msg)
	if !st.OK() {
		t.Fatal(st)
	}
}

func TestChannel_Write__should_write_message_if_it_exceeds_half_window_size(t *testing.T) {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()

	server := testServer(t, func(ch Channel) status.Status {
		<-timer.C

		ch.Read(nil)
		ch.Read(nil)
		return status.OK
	})
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	assert.Equal(t, ch.state.window, ch.state.writeWindow)
	msg0 := bytes.Repeat([]byte("a"), (ch.state.window/2)-100)

	st := ch.Write(nil, msg0)
	if !st.OK() {
		t.Fatal(st)
	}

	msg1 := bytes.Repeat([]byte("a"), (ch.state.window/2)+100)
	st = ch.Write(nil, msg1)
	if !st.OK() {
		t.Fatal(st)
	}
}

// Close

func TestChannel_Close__should_send_close_message(t *testing.T) {
	closed := make(chan struct{})
	server := testServer(t, func(s Channel) status.Status {
		defer close(closed)

		_, st := s.Read(nil)
		return st
	})

	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	st := ch.Write(nil, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	st = ch.Close()
	if !st.OK() {
		t.Fatal(st)
	}

	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestChannel_Close__should_close_incoming_queue(t *testing.T) {
	server := testRequestServer(t)
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	st := ch.Close()
	if !st.OK() {
		t.Fatal(st)
	}

	assert.True(t, ch.state.readQueue.Closed())
}

func TestChannel_Close__should_ignore_when_already_closed(t *testing.T) {
	server := testServer(t, func(s Channel) status.Status {
		_, st := s.Read(nil)
		return st
	})

	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	st := ch.Write(nil, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	st = ch.Close()
	if !st.OK() {
		t.Fatal(st)
	}

	st = ch.Close()
	if !st.OK() {
		t.Fatal(st)
	}
}
