// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"bytes"
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testChannelSend(t *testing.T, ctx async.Context, ch Channel, msg string) {
	t.Helper()

	st := ch.Send(ctx, []byte(msg))
	if !st.OK() {
		t.Fatal(st)
	}
}

// Receive

func TestChannel_Receive__should_receive_message(t *testing.T) {
	server := testServer(t, func(ctx Context, ch Channel) status.Status {
		return ch.Send(ctx, []byte("hello, channel"))
	})

	ctx := async.NoContext()
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	st := ch.Send(ctx, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	msg, st := ch.Receive(ctx)
	if !st.OK() {
		t.Fatal(st)
	}

	assert.Equal(t, []byte("hello, channel"), msg)
}

func TestChannel_Receive__should_return_end_when_channel_closed(t *testing.T) {
	server := testServer(t, func(ctx Context, ch Channel) status.Status {
		return status.OK
	})

	ctx := async.NoContext()
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	st := ch.Send(ctx, nil)
	if !st.OK() {
		t.Fatal(st)
	}

	_, st = ch.Receive(ctx)
	assert.Equal(t, status.End, st)
}

func TestChannel_Receive__should_read_pending_messages_even_when_closed(t *testing.T) {
	sent := make(chan struct{})
	server := testServer(t, func(ctx Context, ch Channel) status.Status {
		defer close(sent)

		st := ch.Send(ctx, []byte("hello, channel"))
		if !st.OK() {
			return st
		}
		return ch.Send(ctx, []byte("how are you?"))
	})

	ctx := async.NoContext()
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	st := ch.Send(ctx, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	select {
	case <-sent:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	msg, st := ch.Receive(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.Equal(t, []byte("hello, channel"), msg)

	msg, st = ch.Receive(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.Equal(t, []byte("how are you?"), msg)

	_, st = ch.Receive(ctx)
	assert.Equal(t, status.End, st)
}

func TestChannel_Receive__should_decrement_recv_window(t *testing.T) {
	server := testServer(t, func(ctx Context, ch Channel) status.Status {
		return ch.Send(ctx, []byte("hello, channel"))
	})

	ctx := async.NoContext()
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	st := ch.Send(ctx, []byte("hello, server"))
	if !st.OK() {
		t.Fatal(st)
	}

	window := ch.unwrap().initWindow
	recvWindow := ch.unwrap().recvWindow.Load()
	require.Equal(t, window, recvWindow)

	_, st = ch.Receive(ctx)
	if !st.OK() {
		t.Fatal(st)
	}

	recvWindow = ch.unwrap().recvWindow.Load()
	assert.Equal(t, len("hello, channel"), window-recvWindow)
}

// Send

func TestChannel_Send__should_send_message(t *testing.T) {
	var msg0 []byte
	done := make(chan struct{})

	server := testServer(t, func(ctx Context, ch Channel) status.Status {
		defer close(done)

		msg, st := ch.Receive(ctx)
		if !st.OK() {
			t.Fatal(st)
		}

		msg0 = bytes.Clone(msg)
		return status.OK
	})

	ctx := async.NoContext()
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	msg1 := []byte("hello, world")
	st := ch.Send(ctx, msg1)
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

func TestChannel_Send__should_send_open_message_if_not_sent(t *testing.T) {
	ctx := async.NoContext()
	server := testRequestServer(t)

	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	opened := ch.unwrap().opened.Load()
	require.False(t, opened)

	st := ch.Send(ctx, []byte("hello, world"))
	if !st.OK() {
		t.Fatal(st)
	}

	opened = ch.unwrap().opened.Load()
	assert.True(t, opened)
}

func TestChannel_Send__should_return_error_when_channel_closed(t *testing.T) {
	ctx := async.NoContext()
	server := testRequestServer(t)

	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()
	ch.SendAndClose(ctx, nil)

	st := ch.Send(ctx, []byte("hello, world"))
	assert.Equal(t, statusChannelClosed, st)
}

func TestChannel_Send__should_decrement_send_window(t *testing.T) {
	ctx := async.NoContext()
	server := testRequestServer(t)

	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	window := ch.unwrap().initWindow
	sendWindow := ch.unwrap().sendWindow.Load()
	require.Equal(t, window, sendWindow)

	st := ch.Send(ctx, []byte("hello, world"))
	if !st.OK() {
		t.Fatal(st)
	}

	sendWindow = ch.unwrap().sendWindow.Load()
	assert.Equal(t, len("hello, world"), window-sendWindow)
}

func TestChannel_Send__should_block_when_send_window_not_enough(t *testing.T) {
	ctx := async.NoContext()
	done := make(chan struct{})
	defer close(done)

	server := testServer(t, func(ctx Context, ch Channel) status.Status {
		<-done
		return status.OK
	})

	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	window := ch.unwrap().initWindow
	sendWindow := ch.unwrap().sendWindow.Load()
	require.Equal(t, window, sendWindow)

	size := int(float64(ch.unwrap().initWindow) / 2.5)
	msg := bytes.Repeat([]byte("a"), size)

	st := ch.Send(ctx, msg)
	if !st.OK() {
		t.Fatal(st)
	}
	st = ch.Send(ctx, msg)
	if !st.OK() {
		t.Fatal(st)
	}

	ctx1 := async.TimeoutContext(time.Millisecond * 100)
	st = ch.Send(ctx1, msg)
	assert.Equal(t, status.Timeout, st)
}

func TestChannel_Send__should_wait_send_window_increment(t *testing.T) {
	server := testServer(t, func(ctx Context, ch Channel) status.Status {
		if _, st := ch.Receive(ctx); !st.OK() {
			return st
		}
		if _, st := ch.Receive(ctx); !st.OK() {
			return st
		}
		return status.OK
	})
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	window := ch.unwrap().initWindow
	sendWindow := ch.unwrap().sendWindow.Load()
	require.Equal(t, window, sendWindow)

	ctx := async.NoContext()
	msg := bytes.Repeat([]byte("a"), int(window/2)+1)

	st := ch.Send(ctx, msg)
	if !st.OK() {
		t.Fatal(st)
	}

	st = ch.Send(ctx, msg)
	if !st.OK() {
		t.Fatal(st)
	}
}

func TestChannel_Send__should_write_message_if_it_exceeds_half_window_size(t *testing.T) {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()

	server := testServer(t, func(ctx Context, ch Channel) status.Status {
		<-timer.C

		if _, st := ch.Receive(ctx); !st.OK() {
			return st
		}
		if _, st := ch.Receive(ctx); !st.OK() {
			return st
		}
		return status.OK
	})
	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	window := ch.unwrap().initWindow
	sendWindow := ch.unwrap().sendWindow.Load()
	require.Equal(t, window, sendWindow)

	ctx := async.NoContext()
	msg0 := bytes.Repeat([]byte("a"), int(window/2)-100)

	st := ch.Send(ctx, msg0)
	if !st.OK() {
		t.Fatal(st)
	}

	msg1 := bytes.Repeat([]byte("a"), int(window/2)+100)
	st = ch.Send(ctx, msg1)
	if !st.OK() {
		t.Fatal(st)
	}
}

// SendAndClose

func TestChannel_SendAndClose__should_send_data_in_close_message(t *testing.T) {
	var msg0 []byte
	done := make(chan struct{})

	server := testServer(t, func(ctx Context, ch Channel) status.Status {
		defer close(done)

		msg, st := ch.Receive(ctx)
		if !st.OK() {
			t.Fatal(st)
		}
		msg0 = append(msg0, msg...)

		msg, st = ch.Receive(ctx)
		if !st.OK() {
			t.Fatal(st)
		}
		msg0 = append(msg0, msg...)
		return status.OK
	})

	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	ctx := async.NoContext()
	st := ch.Send(ctx, []byte("hello, "))
	if !st.OK() {
		t.Fatal(st)
	}

	st = ch.SendAndClose(ctx, []byte("world"))
	if !st.OK() {
		t.Fatal(st)
	}

	closed := ch.unwrap().closed.Load()
	require.True(t, closed)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	assert.Equal(t, []byte("hello, world"), msg0)
}

func TestChannel_SendAndClose__should_send_close_message(t *testing.T) {
	closed := make(chan struct{})
	server := testServer(t, func(ctx Context, ch Channel) status.Status {
		defer close(closed)

		_, st := ch.Receive(ctx)
		return st
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

	st = ch.SendAndClose(ctx, nil)
	if !st.OK() {
		t.Fatal(st)
	}

	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestChannel_SendAndClose__should_close_recv_queue(t *testing.T) {
	ctx := async.NoContext()
	server := testRequestServer(t)

	conn := testConnect(t, server)
	defer conn.Free()

	ch := testChannel(t, conn)
	defer ch.Free()

	st := ch.SendAndClose(ctx, nil)
	if !st.OK() {
		t.Fatal(st)
	}

	assert.True(t, ch.unwrap().recvQueue.Closed())
}

func TestChannel_SendAndClose__should_return_error_when_already_closed(t *testing.T) {
	server := testServer(t, func(ctx Context, ch Channel) status.Status {
		_, st := ch.Receive(ctx)
		return st
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

	st = ch.SendAndClose(ctx, nil)
	if !st.OK() {
		t.Fatal(st)
	}

	st = ch.SendAndClose(ctx, nil)
	assert.Equal(t, statusChannelClosed, st)
}
