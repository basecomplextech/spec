// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"fmt"
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
	sendWindow     atomic.Int32  // remaining send window, can become negative on sending large messages
	sendWindowWait chan struct{} // wait for send window increment

	recvMu     sync.Mutex
	recvQueue  alloc.ByteQueue // data queue
	recvWindow atomic.Int32    // remaining recv window, can become negative on receiving large messages
}

// newChannel2 returns a new outgoing channel.
func newChannel2(c internalConn, client bool, window int32) *channel2 {
	ch := &channel2{}
	ch.id = bin.Random128()
	ch.ctx = newContext(nil) // TODO: conn
	ch.conn = c

	ch.client = client
	ch.initWindow = window

	ch.sendWindow.Store(window)
	ch.sendWindowWait = make(chan struct{}, 1)

	ch.recvWindow.Store(window)
	ch.recvQueue = alloc.NewByteQueue()
	return ch
}

// openChannel2 inits a new incoming channel.
func openChannel2(c internalConn, client bool, msg pmpx.ChannelOpen) *channel2 {
	id := msg.Id()
	window := msg.Window()

	ch := &channel2{}
	ch.id = id
	ch.ctx = newContext(nil) // TODO: conn
	ch.conn = c

	ch.client = client
	ch.initWindow = window

	ch.opened.Store(true)

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
	if ch.opened.Load() {
		return ch.sendData(ctx, data)
	}

	if !ch.client {
		panic("server channel not opened")
	}
	return ch.sendOpen(ctx, data)
}

// SendAndClose sends a close message with a payload.
func (ch *channel2) SendAndClose(ctx async.Context, data []byte) status.Status {
	ch.sendMu.Lock()
	defer ch.sendMu.Unlock()

	if ch.closed.Load() {
		return statusChannelClosed
	}
	if ch.opened.Load() {
		return ch.sendDataClose(ctx, data)
	}

	if !ch.client {
		panic("server channel not opened")
	}
	return ch.sendOpenClose(ctx, data)
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
		case <-ch.ReceiveWait():
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

	delta := ch.incrementRecvWindow()
	if delta > 0 {
		st = ch.sendWindowDelta(ctx, delta)
		if !st.OK() && st != statusChannelClosed {
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
	// Try to close channel
	closed := ch.closed.CompareAndSwap(false, true)
	if !closed {
		return
	}

	// Cancel context, close receive queue
	ch.ctx.Cancel()
	ch.recvQueue.Close()

	// Return if not opened
	if !ch.opened.Load() {
		return
	}

	// Send close message
	st := ch.sendClose(ch.ctx)
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeClosed,
		status.CodeEnd:
	default:
		panic(fmt.Sprintf("unexpected status: %v", st))
	}
}

func (ch *channel2) releaseUser() {

}

func (ch *channel2) sendWindowDelta(ctx async.Context, delta int32) status.Status {
	ch.sendMu.Lock()
	defer ch.sendMu.Unlock()

	if ch.closed.Load() {
		return statusChannelClosed
	}
	if !ch.opened.Load() {
		panic("channel not opened")
	}

	// Build message
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	w := pmpx.NewMessageWriterBuffer(buf)
	msg, err := pmpx.BuildChannelWindow(w, ch.id, delta)
	if err != nil {
		return mpxError(err)
	}

	// Send message
	return ch.conn.send(ctx, msg)
}

// send open

func (ch *channel2) sendOpen(ctx async.Context, data []byte) status.Status {
	if ch.opened.Load() {
		panic("channel already opened")
	}

	// Build batch
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	b := pmpx.NewChannelBatchBuilder(buf, ch.id)
	b, err := b.Open(ch.initWindow)
	if err != nil {
		return mpxError(err)
	}
	b, err = b.Data(data)
	if err != nil {
		return mpxError(err)
	}
	msg, err := b.Build()
	if err != nil {
		return mpxError(err)
	}

	// Decrement window
	size := int32(len(data))
	ch.sendWindow.Add(-size)

	// Send message
	if st := ch.conn.send(ctx, msg); !st.OK() {
		return st
	}

	// Set opened
	ch.opened.Store(true)
	return status.OK
}

func (ch *channel2) sendOpenClose(ctx async.Context, data []byte) status.Status {
	if ch.opened.Load() {
		panic("channel already opened")
	}

	// Build batch
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	b := pmpx.NewChannelBatchBuilder(buf, ch.id)
	b, err := b.Open(ch.initWindow)
	if err != nil {
		return mpxError(err)
	}
	b, err = b.Data(data)
	if err != nil {
		return mpxError(err)
	}
	b, err = b.Close()
	if err != nil {
		return mpxError(err)
	}
	msg, err := b.Build()
	if err != nil {
		return mpxError(err)
	}

	// Decrement window
	size := int32(len(data))
	ch.sendWindow.Add(-size)

	// Send message
	if st := ch.conn.send(ctx, msg); !st.OK() {
		return st
	}

	// Set opened/closed
	ch.opened.Store(true)
	ch.closed.Store(true)
	return status.OK
}

// send close

func (ch *channel2) sendClose(ctx async.Context) status.Status {
	// Build message
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	w := pmpx.NewMessageWriterBuffer(buf)
	msg, err := pmpx.BuildChannelClose(w, ch.id, nil)
	if err != nil {
		return mpxError(err)
	}

	// Send message
	return ch.conn.send(ctx, msg)
}

// send data

func (ch *channel2) sendData(ctx async.Context, data []byte) status.Status {
	// Build message
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	w := pmpx.NewMessageWriterBuffer(buf)
	msg, err := pmpx.BuildChannelData(w, ch.id, data)
	if err != nil {
		return mpxError(err)
	}

	// Decrement window, await increment
	size := int32(len(data))
	if st := ch.decrementSendWindow(ctx, size); !st.OK() {
		return st
	}

	// Send message
	return ch.conn.send(ctx, msg)
}

func (ch *channel2) sendDataClose(ctx async.Context, data []byte) status.Status {
	if ch.closed.Load() {
		panic("channel already closed")
	}

	// Build batch
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	b := pmpx.NewChannelBatchBuilder(buf, ch.id)
	b, err := b.Data(data)
	if err != nil {
		return mpxError(err)
	}
	b, err = b.Close()
	if err != nil {
		return mpxError(err)
	}
	msg, err := b.Build()
	if err != nil {
		return mpxError(err)
	}

	// Decrement window
	size := int32(len(data))
	ch.sendWindow.Add(-size)

	// Send message
	if st := ch.conn.send(ctx, msg); !st.OK() {
		return st
	}

	// Set closed
	ch.closed.Store(true)
	return status.OK
}

// window

func (ch *channel2) decrementSendWindow(ctx async.Context, size int32) status.Status {
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
