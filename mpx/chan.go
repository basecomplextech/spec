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

type Chan interface {
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

type internalChan interface {
	// id returns the channel id.
	id() bin.Bin128

	// closeConn is called by the connection to close the channel.
	closeConn(st status.Status)

	// receiveConn is called by the connection to receive a message.
	receiveConn(msg pmpx.Message) status.Status

	// freeConn is called by the connection to free the channel.
	freeConn()
}

// internal

var (
	_ Chan         = (*channel2)(nil)
	_ internalChan = (*channel2)(nil)
)

type channel2 struct {
	refs  atomic.Int32 // 2 by default (1 for user, 1 for connection)
	state atomic.Pointer[channelState2]
}

// newChannel2 returns a new outgoing channel.
func newChannel2(conn internalConn, client bool, id bin.Bin128, window int32) *channel2 {
	s := newChannelState2(conn, client, id, window)

	ch := &channel2{}
	ch.refs.Store(2)
	ch.state.Store(s)
	return ch
}

// openChannel2 opens and returns a new incoming channel.
func openChannel2(conn internalConn, client bool, msg pmpx.ChannelOpen) *channel2 {
	s := openChannelState2(conn, client, msg)

	ch := &channel2{}
	ch.refs.Store(2)
	ch.state.Store(s)
	return ch
}

// Context returns a channel context.
func (ch *channel2) Context() Context {
	s := ch.acquire()
	defer ch.release()

	return s.ctx
}

// Send

// Send sends a message to the channel.
func (ch *channel2) Send(ctx async.Context, data []byte) status.Status {
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
	return s.sender.sendOpenData(ctx, data)
}

// SendAndClose sends a close message with a payload.
func (ch *channel2) SendAndClose(ctx async.Context, data []byte) status.Status {
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
		return s.sender.sendDataClose(ctx, data)
	}

	// Open/close channel
	s.open()
	s.close()

	// Decrement window
	size := int32(len(data))
	s.sendWindow.Add(-size)

	// Send open/data/close
	return s.sender.sendOpenDataClose(ctx, data)
}

// Receive

// Receive receives and returns a message, or an end status.
func (ch *channel2) Receive(ctx async.Context) ([]byte, status.Status) {
	s := ch.acquire()
	defer ch.release()

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
		case <-s.recvQueue.ReadWait():
		}
	}
}

// ReceiveAsync receives and returns a message, or false/end.
//
// The message is valid until the next call to Receive.
// The method does not block if no messages, and returns false instead.
func (ch *channel2) ReceiveAsync(ctx async.Context) ([]byte, bool, status.Status) {
	s := ch.acquire()
	defer ch.release()

	// Read next message
	data, ok, st := s.recvQueue.Read()
	if !ok || !st.OK() {
		return nil, ok, st
	}

	// Try to increment window
	delta := s.incrementRecvWindow()
	if delta <= 0 {
		return data, true, status.OK
	}

	// Send window increment
	if !s.closed.Load() {
		st := s.sender.sendWindow(ctx, delta)
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
func (ch *channel2) ReceiveWait() <-chan struct{} {
	s := ch.acquire()
	defer ch.release()

	return s.recvQueue.ReadWait()
}

// Internal

// Free closes the channel and releases its resources.
func (ch *channel2) Free() {
	ch.closeUser()
	ch.release()
}

// internal

// id returns the channel id.
func (ch *channel2) id() bin.Bin128 {
	s := ch.acquire()
	defer ch.release()

	return s.id
}

// closeConn is called by the connection to close the channel.
func (ch *channel2) closeConn(st status.Status) {
	s := ch.acquire()
	defer ch.release()

	s.close()
}

// receiveConn is called by the connection to receive a message.
func (ch *channel2) receiveConn(msg pmpx.Message) status.Status {
	s := ch.acquire()
	defer ch.release()

	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	// Ignore messages if closed
	if s.closed.Load() {
		return status.OK
	}

	// Receive message
	return s.receiveMessage(msg, false /* not inside batch */)
}

// freeConn is called by the connection to free the channel.
func (ch *channel2) freeConn() {
	ch.closeConn(statusConnClosed)
	ch.release()
}

// acquire/release

// acquire increments the refcounter and returns the channel state, panics if freed.
func (ch *channel2) acquire() *channelState2 {
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
func (ch *channel2) release() {
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

func (ch *channel2) closeUser() {
	s := ch.acquire()
	defer ch.release()

	// Close user once
	closed := s.closedUser.CompareAndSwap(false, true)
	if !closed {
		return
	}

	// Close channel in defer
	// We need context to send message.
	defer s.close()

	// Send close message
	st := s.sender.sendClose(s.ctx)
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeClosed,
		status.CodeEnd:
	default:
		panic(fmt.Sprintf("unexpected status: %v", st)) // unreachable
	}
}
