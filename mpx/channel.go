// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"fmt"
	"sync/atomic"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/pmpx"
)

type Channel interface {
	// Conn returns a channel connection.
	Conn() Conn

	// Context returns a channel context.
	Context() Context

	// Send

	// Send sends a message to the channel.
	Send(ctx async.Context, data []byte) status.Status

	// SendAndClose sends a close message with a payload.
	SendAndClose(ctx async.Context, data []byte) status.Status

	// Receive

	// Receive receives and returns a message, or an end status.
	//
	// The message is valid until the next call to Receive.
	// The method blocks until a message is received, or the channel is closed.
	Receive(ctx async.Context) ([]byte, status.Status)

	// ReceiveAsync receives and returns a message, or false/end.
	//
	// The message is valid until the next call to Receive.
	// The method does not block if no messages, and returns false instead.
	ReceiveAsync(ctx async.Context) ([]byte, bool, status.Status)

	// ReceiveWait returns a channel that is notified on a new message, or a channel close.
	ReceiveWait() <-chan struct{}

	// Internal

	// Free closes the channel and releases its resources.
	Free()
}

// internal

type internalChannel interface {
	// receive is called by the connection to receive a message.
	receive(msg pmpx.Message) status.Status

	// free is called by the connection to free the channel.
	free()
}

// implementation

var (
	_ Channel         = (*channel)(nil)
	_ internalChannel = (*channel)(nil)
)

type channel struct {
	refs  atomic.Int32 // 2 by default (1 for user, 1 for connection)
	freed atomic.Bool  // ensures public free is called once

	state atomic.Pointer[channelState]
}

// newChannel returns a new outgoing channel.
func newChannel(conn internalConn, client bool, id bin.Bin128, window int32) *channel {
	s := newChannelState(conn, client, id, window)

	ch := &channel{}
	ch.refs.Store(2)
	ch.state.Store(s)
	return ch
}

// openChannel opens and returns a new incoming channel.
func openChannel(conn internalConn, client bool, msg pmpx.ChannelOpen) *channel {
	s := openChannelState(conn, client, msg)

	// Make channel
	ch := &channel{}
	ch.refs.Store(2)
	ch.state.Store(s)

	// Maybe receive data
	data := msg.Data()
	if len(data) > 0 {
		_, _ = s.recvQueue.Write(data) // receive queue is unbounded, ignore status
	}
	return ch
}

// Conn returns a channel connection.
func (ch *channel) Conn() Conn {
	s := ch.acquire()
	defer ch.release()

	return s.conn
}

// Context returns a channel context.
func (ch *channel) Context() Context {
	s := ch.acquire()
	defer ch.release()

	return s.ctx
}

// Send

// Send sends a message to the channel.
func (ch *channel) Send(ctx async.Context, data []byte) status.Status {
	s := ch.acquire()
	defer ch.release()

	// Lock send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	if s.closed.Load() {
		return statusChannelClosed
	}

	// If opened, send data
	if s.opened.Load() {
		// Decrement window, await
		if st := s.decrementSendWindow(ctx, data); !st.OK() {
			return st
		}
		// Send message
		return s.sender.sendData(ctx, data)
	}

	// Open channel
	s.open()

	// Decrement window
	size := int32(len(data))
	s.sendWindow.Add(-size)

	// Send open/data
	return s.sender.sendOpen(ctx, data)
}

// SendAndClose sends a close message with a payload.
func (ch *channel) SendAndClose(ctx async.Context, data []byte) status.Status {
	s := ch.acquire()
	defer ch.release()

	// Lock send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	if s.closed.Load() {
		return statusChannelClosed
	}

	// If opened, close, send data/close
	if s.opened.Load() {
		s.close()

		// Decrement window
		size := int32(len(data))
		s.sendWindow.Add(-size)

		// Send message
		return s.sender.sendClose(ctx, data)
	}

	// Open/close channel
	s.open()
	s.close()

	// Decrement window
	size := int32(len(data))
	s.sendWindow.Add(-size)

	// Send open/data/close
	return s.sender.sendOpenClose(ctx, data)
}

// Receive

// Receive receives and returns a message, or an end status.
func (ch *channel) Receive(ctx async.Context) ([]byte, status.Status) {
	for {
		// Poll channel
		data, ok, st := ch.ReceiveAsync(ctx)
		switch {
		case !st.OK():
			return nil, st
		case ok:
			return data, status.OK
		}

		// Await new message or close
		select {
		case <-ctx.Wait():
			return nil, ctx.Status()
		case <-ch.ReceiveWait():
		}
	}
}

// ReceiveAsync receives and returns a message, or false/end.
//
// The message is valid until the next call to Receive.
// The method does not block if no messages, and returns false instead.
func (ch *channel) ReceiveAsync(ctx async.Context) ([]byte, bool, status.Status) {
	s := ch.acquire()
	defer ch.release()

	// Read next message
	data, ok, st := s.recvQueue.Read()
	if !ok || !st.OK() {
		return nil, ok, st
	}

	// Increment received
	size := int32(len(data))
	recv := s.recvBytes.Add(size)

	// Check window/2 reached
	if recv < s.initWindow/2 {
		return data, true, status.OK
	}

	// Decrement bytes, send window delta
	s.recvBytes.Add(-recv)
	if !s.closed.Load() {
		st := s.sender.sendWindow(ctx, recv)
		switch st.Code {
		case status.CodeOK,
			status.CodeCancelled,
			status.CodeClosed,
			status.CodeEnd:
		default:
			return nil, false, st // unreachable
		}
	}
	return data, true, status.OK
}

// ReceiveWait returns a channel that is notified on a new message, or a channel close.
func (ch *channel) ReceiveWait() <-chan struct{} {
	s := ch.acquire()
	defer ch.release()

	return s.recvQueue.ReadWait()
}

// Internal

// Free closes the channel and releases its resources.
func (ch *channel) Free() {
	ok := ch.freed.CompareAndSwap(false, true)
	if !ok {
		panic("free called multiple times")
	}

	ch.closeUser()
	ch.release()
}

// internal

// receive is called by the connection to receive a message.
func (ch *channel) receive(msg pmpx.Message) status.Status {
	s := ch.acquire()
	defer ch.release()

	// Ignore messages if closed
	if s.closed.Load() {
		return status.OK
	}

	// Receive message
	return s.receiveMessage(msg)
}

// free is called by the connection to free the channel.
func (ch *channel) free() {
	s := ch.state.Load()
	if s == nil {
		panic("free of freed channel")
	}

	defer ch.release()
	s.close()
}

// acquire/release

// acquire increments the refcounter and returns the channel state, panics if freed.
func (ch *channel) acquire() *channelState {
	refs := ch.refs.Add(1)
	if refs == 1 {
		panic("acquire of freed channel")
	}

	s := ch.state.Load()
	if s == nil {
		panic("acquire of freed channel")
	}
	return s
}

// release decrements the internal refs counter.
func (ch *channel) release() {
	refs := ch.refs.Add(-1)
	if refs > 0 {
		return
	}

	s := ch.state.Swap(nil)
	if s == nil {
		panic("release of released channel")
	}
	releaseChannelState2(s)
}

// close

func (ch *channel) closeUser() {
	s := ch.acquire()
	defer ch.release()

	// Check already closed
	closed := s.closed.Load()
	if closed {
		return
	}

	// Close channel in defer
	// We need context to send message.
	defer s.close()

	// Send close message
	st := s.sender.sendClose(s.ctx, nil /* no data */)
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeClosed,
		status.CodeEnd:
	default:
		panic(fmt.Sprintf("unexpected status: %v", st)) // unreachable
	}
}

// unwrap

// unwrap returns the channel state, used in tests.
func (ch *channel) unwrap() *channelState {
	return ch.state.Load()
}
