// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/pools"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/pmpx"
)

type Channel interface {
	// Context returns a channel context.
	Context() Context

	// Receive

	// Receive receives and returns a message, or an end.
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

	// Send

	// Send sends a message to the channel.
	Send(ctx async.Context, msg []byte) status.Status

	// SendAndClose sends a close message with a payload.
	SendAndClose(ctx async.Context, msg []byte) status.Status

	// SendClose sends a close message, and closes the channel.
	//
	// No more messages can be sent after this call.
	SendClose(ctx async.Context) status.Status

	// Internal

	// Free closes the channel and releases its resources.
	Free()
}

// internal

type internalChannel interface {
	// Free1 is called when the channel is freed by the connection.
	Free1()

	// Receive1 is called when the channel receives a message from the connection.
	Receive1(msg pmpx.Message) status.Status
}

var (
	_ Channel         = (*channel)(nil)
	_ internalChannel = (*channel)(nil)
)

type channel struct {
	stateMu sync.RWMutex
	state   *channelState
}

type channelState struct {
	id     bin.Bin128
	ctx    Context
	conn   conn
	client bool // client or server channel
	window int  // initial window size

	// send
	sendMu      sync.Mutex
	sendOpen    bool // open message sent
	sendClose   bool // close message sent
	sendFree    bool // freed by user
	sendBuilder builder

	// receive
	recvMu    sync.Mutex
	recvOpen  bool // open message received
	recvClose bool // close message received
	recvFree  bool // freed by connection
	recvQueue alloc.ByteQueue

	// windows
	windowMu   sync.Mutex
	windowSend int           // remaining send window, can become negative on sending large messages
	windowRecv int           // remaining recv window, can become negative on receiving large messages
	windowWait chan struct{} // wait for send window increment
}

// createChannel creates a new outgoing channel.
func createChannel(c conn, client bool, id bin.Bin128, window int) *channel {
	s := acquireChannelState()
	s.id = id
	s.ctx = newContext(c)
	s.conn = c
	s.client = client
	s.window = window

	s.recvOpen = true

	s.windowRecv = window
	s.windowSend = window

	return &channel{state: s}
}

// openChannel inits a new incoming channel.
func openChannel(c conn, client bool, msg pmpx.ChannelOpen) *channel {
	id := msg.Id()
	window := int(msg.Window())

	s := acquireChannelState()
	s.id = id
	s.ctx = newContext(c)
	s.conn = c
	s.client = client
	s.window = window

	s.sendOpen = true

	s.windowRecv = window
	s.windowSend = window

	return &channel{state: s}
}

func newChannelState() *channelState {
	return &channelState{
		sendBuilder: newBuilder(),
		recvQueue:   alloc.NewByteQueue(),
		windowWait:  make(chan struct{}, 1),
	}
}

// Context returns a channel context.
func (ch *channel) Context() Context {
	s, ok := ch.rlock()
	if !ok {
		return closedContext
	}
	defer ch.stateMu.RUnlock()

	return s.ctx
}

// Receive

// Receive receives and returns a message, or an end.
//
// The message is valid until the next call to Receive.
// The method blocks until a message is received, or the channel is closed.
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
	data, ok, st := ch.receive()
	if !ok || !st.OK() {
		return nil, false, st
	}

	delta := ch.incrementRecvWindow()
	if delta == 0 {
		return data, true, status.OK
	}

	st = ch.sendWindow(delta)
	if !st.OK() && st != statusChannelClosed {
		return nil, false, st
	}
	return data, true, status.OK
}

// ReceiveWait returns a channel that is notified on a new message, or a channel close.
func (ch *channel) ReceiveWait() <-chan struct{} {
	s, ok := ch.rlock()
	if !ok {
		return closedChan
	}
	defer ch.stateMu.RUnlock()

	return s.recvQueue.ReadWait()
}

// Send

// Send sends a message to the channel.
func (ch *channel) Send(ctx async.Context, data []byte) status.Status {
	input := messageInput{
		data: data,
	}
	return ch.sendMessage(ctx, input)
}

// SendAndClose sends a close message with a payload.
func (ch *channel) SendAndClose(ctx async.Context, data []byte) status.Status {
	input := messageInput{
		data:  data,
		close: true,
	}
	if st := ch.sendMessage(ctx, input); !st.OK() {
		return st
	}

	// Call again in case the previous message was an open message
	return ch.sendClose(ctx)
}

// SendClose sends a close message, and closes the channel.
func (ch *channel) SendClose(ctx async.Context) status.Status {
	return ch.sendClose(ctx)
}

// Internal

// Free closes the channel and releases its resources.
func (ch *channel) Free() {
	_ = ch.sendClose(nil /* use channel context */) // ignore closed status
	ch.freeSend()
	ch.free()
}

// Free1 is called when the channel is freed by the connection.
func (ch *channel) Free1() {
	ch.freeReceive()
	ch.free()
}

// Receive1 is called when a message is received.
func (ch *channel) Receive1(msg pmpx.Message) status.Status {
	code := msg.Code()

	switch code {
	case pmpx.Code_ChannelOpen:
		return ch.receiveOpen(msg.ChannelOpen())
	case pmpx.Code_ChannelClose:
		return ch.receiveClose(msg.ChannelClose())
	case pmpx.Code_ChannelWindow:
		return ch.receiveWindow(msg.ChannelWindow())
	case pmpx.Code_ChannelData:
		return ch.receiveData(msg.ChannelData())
	}

	return mpxErrorf("received unexpected message, code=%v", code)
}

// private

func (ch *channel) free() {
	ch.stateMu.Lock()
	defer ch.stateMu.Unlock()

	s := ch.state
	if s == nil {
		return
	}

	if !s.sendFree || !s.recvFree {
		return
	}

	ch.state = nil
	releaseChannelState(s)
}

func (ch *channel) rlock() (*channelState, bool) {
	ch.stateMu.RLock()
	s := ch.state
	if s == nil {
		ch.stateMu.RUnlock()
		return nil, false
	}
	return s, true
}

// receive

func (ch *channel) receive() ([]byte, bool, status.Status) {
	// RLock state
	s, ok := ch.rlock()
	if !ok {
		return nil, false, statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Poll queue
	data, ok, st := s.recvQueue.Read()
	if !ok || !st.OK() {
		return nil, false, st
	}

	// Decrement recv window
	s.decrementRecvWindow(len(data))
	return data, true, status.OK
}

// send

func (ch *channel) sendMessage(ctx async.Context, input messageInput) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Decrement window
	size := len(input.data)
	if st := s.decrementSendWindow(ctx, size, true /* wait */); !st.OK() {
		return st
	}

	// Lock send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	if s.sendClose {
		return statusChannelClosed
	}

	// Make message
	input.id = s.id
	input.open = !s.sendOpen
	input.window = int32(s.window)
	if input.open {
		input.close = false
	}

	buf := alloc.AcquireBuffer()
	defer buf.Free()

	msg, err := s.sendBuilder.buildMessage(buf, input)
	if err != nil {
		return mpxError(err)
	}

	// Write message
	if st := s.conn.SendMessage(ctx, msg); !st.OK() {
		return st
	}

	// Update state
	s.sendOpen = true
	s.sendClose = s.sendClose || input.close
	return status.OK
}

func (ch *channel) sendClose(ctxOrNil async.Context) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Context is nil in Free
	ctx := ctxOrNil
	if ctx == nil {
		ctx = s.ctx
	}

	// Lock send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	switch {
	case s.sendClose:
		return status.OK
	case !s.sendOpen:
		s.sendClose = true
		return status.OK
	}

	// Make message
	input := messageInput{
		id:    s.id,
		close: true,
	}

	buf := alloc.AcquireBuffer()
	defer buf.Free()

	msg, err := s.sendBuilder.buildMessage(buf, input)
	if err != nil {
		return mpxError(err)
	}

	// Use channel context to send close message
	ctx1 := s.ctx

	// Decrement window
	n := len(msg.Unwrap().Raw())
	if st := s.decrementSendWindow(ctx1, n, false /* no wait, force */); !st.OK() {
		return st
	}

	// Write message
	if st := s.conn.SendMessage(ctx1, msg); !st.OK() {
		return st
	}

	// Update state
	s.sendClose = true
	return status.OK
}

func (ch *channel) sendWindow(delta int32) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Lock send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	// Make message
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	msg, err := s.sendBuilder.buildWindow(buf, s.id, delta)
	if err != nil {
		return mpxError(err)
	}

	// Write message
	ctx := async.NoContext()
	return s.conn.SendMessage(ctx, msg)
}

func (ch *channel) freeSend() {
	s, ok := ch.rlock()
	if !ok {
		return
	}
	defer ch.stateMu.RUnlock()

	// Lock send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	if s.sendFree {
		return
	}

	if !s.sendClose {
		panic("cannot free open cannel")
	}
	s.sendFree = true
}

// internal receive

func (ch *channel) receiveOpen(msg pmpx.ChannelOpen) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Check server
	if s.client {
		return mpxErrorf("received open message on client channel, channel=%v", s.id)
	}

	// Lock send/receive
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if s.recvOpen {
		return mpxErrorf("received duplicate open message, channel=%v", s.id)
	}

	// Handle open
	data := msg.Data()
	s.recvOpen = true

	// Maybe write data
	if len(data) != 0 {
		_, _ = s.recvQueue.Write(data) // ignore end and false, read queues are unbounded
	}
	return status.OK
}

func (ch *channel) receiveClose(msg pmpx.ChannelClose) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Lock receive
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if s.recvClose {
		return mpxErrorf("received duplicate close message, channel=%v", s.id)
	}
	s.recvClose = true

	// Maybe write data
	data := msg.Data()
	if len(data) != 0 {
		_, _ = s.recvQueue.Write(data) // ignore end and false, read queues are unbounded
	}

	// Cancel context, close queue
	s.ctx.Cancel()
	s.recvQueue.Close()
	return status.OK
}

func (ch *channel) receiveData(data pmpx.ChannelData) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Lock receive
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if s.recvClose {
		return mpxErrorf("received message after close, channel=%v", s.id)
	}

	// Write data
	b := data.Data()
	_, _ = s.recvQueue.Write(b) // ignore end and false, read queues are unbounded
	return status.OK
}

func (ch *channel) receiveWindow(msg pmpx.ChannelWindow) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Lock receive
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if s.recvClose {
		return mpxErrorf("received window after close, channel=%v", s.id)
	}

	// Increment window
	delta := int(msg.Delta())
	s.incrementSendWindow(delta)
	return status.OK
}

func (ch *channel) freeReceive() {
	s, ok := ch.rlock()
	if !ok {
		return
	}
	defer ch.stateMu.RUnlock()

	// Lock send/receive
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if s.recvFree {
		return
	}

	// Close send
	s.sendClose = true

	// Close/free receive
	s.recvClose = true
	s.recvFree = true
	s.recvQueue.Close()

	// Cancel context
	s.ctx.Cancel()
}

// send window

func (s *channelState) incrementSendWindow(delta int) {
	s.windowMu.Lock()
	defer s.windowMu.Unlock()

	s.windowSend += delta

	select {
	case s.windowWait <- struct{}{}:
	default:
	}
}

func (s *channelState) decrementSendWindow(ctx async.Context, n int, wait bool) status.Status {
	for {
		// Decrement send window for normal small messages
		s.windowMu.Lock()
		if !wait || s.windowSend >= n {
			s.windowSend -= n
			s.windowMu.Unlock()
			return status.OK
		}

		// Decrement send window for large messages, when the remaining window
		// is greater than the half of the initial window, but the message size
		// still exceeds it.
		if s.windowSend >= s.window/2 {
			s.windowSend -= n
			s.windowMu.Unlock()
			return status.OK
		}
		s.windowMu.Unlock()

		// Wait for send window increment
		select {
		case <-ctx.Wait():
			return ctx.Status()
		case <-s.ctx.Wait():
			return statusChannelClosed
		case <-s.windowWait:
		}
	}
}

// recv window

func (ch *channel) incrementRecvWindow() int32 {
	s, ok := ch.rlock()
	if !ok {
		return 0
	}
	defer ch.stateMu.RUnlock()

	// Lock window
	s.windowMu.Lock()
	defer s.windowMu.Unlock()

	if s.windowRecv > s.window/2 {
		return 0
	}

	// Increment recv window
	delta := s.window - s.windowRecv
	s.windowRecv += delta
	return int32(delta)
}

func (s *channelState) decrementRecvWindow(size int) {
	s.windowMu.Lock()
	defer s.windowMu.Unlock()

	s.windowRecv -= size
}

// state pool

var channelPool = pools.NewPoolFunc(newChannelState)

func acquireChannelState() *channelState {
	return channelPool.New()
}

func releaseChannelState(s *channelState) {
	s.reset()
	channelPool.Put(s)
}

func (s *channelState) reset() {
	s.id = bin.Bin128{}
	s.conn = nil
	s.client = false
	s.window = 0

	s.ctx.Free()
	s.ctx = nil

	// Reset send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	s.sendOpen = false
	s.sendClose = false
	s.sendFree = false

	// Reset receive
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	s.recvOpen = false
	s.recvClose = false
	s.recvFree = false
	s.recvQueue.Reset()

	// Reset windows
	s.windowMu.Lock()
	defer s.windowMu.Unlock()

	s.windowSend = 0
	s.windowRecv = 0
	select {
	case <-s.windowWait:
	default:
	}
}
