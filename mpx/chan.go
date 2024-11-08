// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"

	"github.com/basecomplextech/baselibrary/alloc"
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

type channel2 struct {
	id   bin.Bin128
	ctx  Context
	conn internalConn

	client     bool  // client or server
	initWindow int32 // initial window size

	opened atomic.Bool
	closed atomic.Bool

	sendMu         sync.Mutex
	sender         chanSender
	sendWindow     atomic.Int32  // remaining send window, can become negative on sending large messages
	sendWindowWait chan struct{} // wait for send window increment

	recvMu     sync.Mutex
	recvQueue  alloc.ByteQueue // data queue
	recvWindow atomic.Int32    // remaining recv window, can become negative on receiving large messages
}

// newChannel2 returns a new outgoing channel.
func newChannel2(conn internalConn, client bool, window int32) *channel2 {
	ch := &channel2{}
	ch.id = bin.Random128()
	ch.ctx = newContext(nil) // TODO: conn
	ch.conn = conn

	ch.client = client
	ch.initWindow = window

	ch.sender = newChanSender(ch, conn)
	ch.sendWindow.Store(window)
	ch.sendWindowWait = make(chan struct{}, 1)

	ch.recvWindow.Store(window)
	ch.recvQueue = alloc.NewByteQueue()
	return ch
}

// openChannel2 inits a new incoming channel.
func openChannel2(conn internalConn, client bool, msg pmpx.ChannelOpen) *channel2 {
	id := msg.Id()
	window := msg.Window()

	ch := &channel2{}
	ch.id = id
	ch.ctx = newContext(nil) // TODO: conn
	ch.conn = conn

	ch.client = client
	ch.initWindow = window

	ch.opened.Store(true)

	ch.sender = newChanSender(ch, conn)
	ch.sendWindow.Store(window)
	ch.sendWindowWait = make(chan struct{}, 1)

	ch.recvWindow.Store(window)
	ch.recvQueue = alloc.NewByteQueue()
	return ch
}

// Context returns a channel context.
func (ch *channel2) Context() Context {
	return ch.ctx
}

// Send

// Send sends a message to the channel.
func (ch *channel2) Send(ctx async.Context, data []byte) status.Status {
	ch.sendMu.Lock()
	defer ch.sendMu.Unlock()

	if ch.closed.Load() {
		return statusChannelClosed
	}

	// If opened, send data
	if ch.opened.Load() {
		// Decrement window, await
		if st := ch.decrementSendWindow(ctx, data); !st.OK() {
			return st
		}

		// Send message
		return ch.sender.sendData(ctx, data)
	}

	// Open channel
	ch.open()

	// Decrement window
	size := int32(len(data))
	ch.sendWindow.Add(-size)

	// Send message
	return ch.sender.sendOpenData(ctx, data)
}

// SendAndClose sends a close message with a payload.
func (ch *channel2) SendAndClose(ctx async.Context, data []byte) status.Status {
	ch.sendMu.Lock()
	defer ch.sendMu.Unlock()

	if ch.closed.Load() {
		return statusChannelClosed
	}

	// If opened, close, send message
	if ch.opened.Load() {
		ch.close()

		// Decrement window
		size := int32(len(data))
		ch.sendWindow.Add(-size)

		// Send message
		return ch.sender.sendDataClose(ctx, data)
	}

	// Open/close channel
	ch.open()
	ch.close()

	// Decrement window
	size := int32(len(data))
	ch.sendWindow.Add(-size)

	// Send message
	return ch.sender.sendOpenDataClose(ctx, data)
}

// Receive

// Receive receives and returns a message, or an end status.
func (ch *channel2) Receive(ctx async.Context) ([]byte, status.Status) {
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
		case <-ch.recvQueue.ReadWait():
		}
	}
}

// ReceiveAsync receives and returns a message, or false/end.
//
// The message is valid until the next call to Receive.
// The method does not block if no messages, and returns false instead.
func (ch *channel2) ReceiveAsync(ctx async.Context) ([]byte, bool, status.Status) {
	data, ok, st := ch.recvQueue.Read()
	if !ok || !st.OK() {
		return nil, ok, st
	}

	// Try to increment window
	delta := ch.incrementRecvWindow()
	if delta <= 0 {
		return data, true, status.OK
	}

	// Send window increment
	if !ch.closed.Load() {
		if st := ch.sender.sendWindow(ctx, delta); !st.OK() {
			return nil, false, st
		}
	}
	return data, true, status.OK
}

// ReceiveWait returns a channel that is notified on a new message, or a channel close.
func (ch *channel2) ReceiveWait() <-chan struct{} {
	return ch.recvQueue.ReadWait()
}

// Internal

// Free closes the channel and releases its resources.
func (ch *channel2) Free() {
	ch.closeUser()
	ch.releaseUser()
}

// internal

func (ch *channel2) closeUser() {
	// Close channel
	closed := ch.close()
	if !closed {
		return
	}

	// Send close message
	st := ch.sender.sendClose(ch.ctx)
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeClosed,
		status.CodeEnd:
		return
	}

	// Unreachable, but still panic
	panic(fmt.Sprintf("unexpected status: %v", st))
}

func (ch *channel2) releaseUser() {

}

// open/close

func (ch *channel2) open() {
	if ch.opened.Load() {
		return
	}
	if !ch.client {
		panic("cannot open server channel")
	}

	ch.opened.Store(true)
}

func (ch *channel2) close() bool {
	// Try to close channel
	closed := ch.closed.CompareAndSwap(false, true)
	if !closed {
		return false
	}

	// Cancel context, close receive queue
	ch.ctx.Cancel()
	ch.recvQueue.Close()
	return true
}

// window

func (ch *channel2) incrementRecvWindow() int32 {
	window := ch.recvWindow.Load()
	if window > ch.initWindow/2 {
		return 0
	}

	// Increment recv window
	delta := ch.initWindow - window
	ch.recvWindow.Add(delta)
	return delta
}

func (ch *channel2) decrementSendWindow(ctx async.Context, data []byte) status.Status {
	// Check data size
	size := len(data)
	if size > math.MaxInt32 {
		return mpxErrorf("message too large, size=%d", size)
	}

	for {
		// Try to decrement window
		window := ch.sendWindow.Load()
		if window >= int32(size) {
			ch.sendWindow.Add(-int32(size))
			return status.OK
		}

		// Await window increment
		select {
		case <-ctx.Wait():
			return ctx.Status()
		case <-ch.ctx.Wait():
			return statusChannelClosed
		case <-ch.sendWindowWait:
		}
	}
}
