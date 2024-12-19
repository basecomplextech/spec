// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package rpc

import (
	"sync"
	"sync/atomic"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/pools"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/mpx"
	"github.com/basecomplextech/spec/proto/prpc"
)

// ServerChannel is a server RPC channel.
type ServerChannel interface {
	// Context returns a channel context.
	Context() Context

	// Receive

	// Request returns the request, the message is valid until the next call to Receive.
	Request(ctx async.Context) (prpc.Request, status.Status)

	// Receive receives and returns a message from the channel, or an end.
	//
	// The method blocks until a message is received, or the channel is closed.
	// The message is valid until the next call to Receive/ReceiveAsync.
	Receive(ctx async.Context) ([]byte, status.Status)

	// ReceiveAsync receives and returns a message, or false/end.
	//
	// The method does not block if no messages, and returns false instead.
	// The message is valid until the next call to Receive/ReceiveAsync.
	ReceiveAsync(ctx async.Context) ([]byte, bool, status.Status)

	// ReceiveWait returns a channel which is notified on a new message, or a channel close.
	ReceiveWait() <-chan struct{}

	// Send

	// Send sends a message to the channel.
	Send(ctx async.Context, message []byte) status.Status

	// SendEnd sends an end message to the channel.
	SendEnd(ctx async.Context) status.Status
}

// internal

type internalServerChannel interface {
	// Method returns a joined method name, the string is valid until the channel is freed.
	Method() string

	// SendResponse sends a response and closes the channel.
	SendResponse(ctx async.Context, result []byte, st status.Status) status.Status
}

var _ ServerChannel = (*serverChannel)(nil)

type serverChannel struct {
	refs  ref.Atomic32 // state refcounter
	freed atomic.Bool  // ensures the channel is freed only once
	state atomic.Pointer[serverChannelState]
}

type serverChannelState struct {
	ch     mpx.Channel // unowned
	method []byte      // call method names, separated by '/'

	// send
	sendMu      sync.Mutex
	sendReq     bool // request sent
	sendEnd     bool // end sent
	sendBuilder builder

	// receive
	recvMu     sync.Mutex
	recvReq    prpc.Request
	recvEnd    bool // end received
	recvFailed bool
	recvError  status.Status
}

func newServerChannel(ch mpx.Channel, req prpc.Request) *serverChannel {
	s := acquireServerState()
	s.ch = ch
	s.method = requestMethod(s.method, req)
	s.recvReq = req

	c := &serverChannel{}
	c.refs.Init(1)
	c.state.Store(s)
	return c
}

func newServerChannelState() *serverChannelState {
	return &serverChannelState{
		sendBuilder: newBuilder(),
	}
}

// Method returns a joined method name, the string is valid until the channel is freed.
func (ch *serverChannel) Method() string {
	s, ok := ch.acquire()
	if !ok {
		return ""
	}
	defer ch.release()

	return unsafeString(s.method)
}

// Context returns a channel context.
func (ch *serverChannel) Context() Context {
	s, ok := ch.acquire()
	if !ok {
		return mpx.ClosedContext()
	}
	defer ch.release()

	return s.ch.Context()
}

// Receive

// Request returns the request, the message is valid until the next call to ReadSync.
func (ch *serverChannel) Request(ctx async.Context) (prpc.Request, status.Status) {
	s, ok := ch.acquire()
	if !ok {
		return prpc.Request{}, status.Closed
	}
	defer ch.release()

	// Lock receive
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	// Check state
	if s.recvReq.IsEmpty() {
		return prpc.Request{}, Error("request already received")
	}

	return s.recvReq, status.OK
}

// Receive receives and returns a message from the channel, or an end.
//
// The method blocks until a message is received, or the channel is closed.
// The message is valid until the next call to Receive/ReceiveAsync.
func (ch *serverChannel) Receive(ctx async.Context) ([]byte, status.Status) {
	for {
		msg, ok, st := ch.ReceiveAsync(ctx)
		switch {
		case !st.OK():
			return nil, st
		case ok:
			return msg, status.OK
		}

		select {
		case <-ctx.Wait():
			return nil, ctx.Status()
		case <-ch.ReceiveWait():
		}
	}
}

// ReceiveAsync receives and returns a message, or false/end.
//
// The method does not block if no messages, and returns false instead.
// The message is valid until the next call to Receive/ReceiveAsync.
func (ch *serverChannel) ReceiveAsync(ctx async.Context) ([]byte, bool, status.Status) {
	s, ok := ch.acquire()
	if !ok {
		return nil, false, status.Closed
	}
	defer ch.release()

	// Lock receive
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	// Check state
	switch {
	case s.recvFailed:
		return nil, false, s.recvError
	case s.recvEnd:
		return nil, false, status.End
	}

	// Clear request
	if !s.recvReq.IsEmpty() {
		s.recvReq = prpc.Request{}
	}

	// Read message
	var msg prpc.Message
	{
		b, ok, st := s.ch.ReceiveAsync(ctx)
		if !ok || !st.OK() {
			return nil, false, st
		}

		var err error
		msg, _, err = prpc.ParseMessage(b)
		if err != nil {
			return nil, false, WrapError(err)
		}
	}

	// Handle message
	typ := msg.Type()
	switch typ {
	case prpc.MessageType_Message:
		return msg.Msg(), true, status.OK

	case prpc.MessageType_End:
		s.recvEnd = true
		return nil, false, status.End
	}

	st := Errorf("unexpected message type %d", typ)
	s.readFail(st)
	return nil, false, st
}

// ReceiveWait returns a channel which is notified on a new message, or a channel close.
func (ch *serverChannel) ReceiveWait() <-chan struct{} {
	s, ok := ch.acquire()
	if !ok {
		return closedChan
	}
	defer ch.release()

	return s.ch.ReceiveWait()
}

// Send

// Send sends a message to the channel.
func (ch *serverChannel) Send(ctx async.Context, message []byte) status.Status {
	s, ok := ch.acquire()
	if !ok {
		return status.Closed
	}
	defer ch.release()

	// Lock send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	if s.sendEnd {
		return Error("end already sent")
	}

	// Make message
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	msg, err := s.sendBuilder.buildMessage(buf, message)
	if err != nil {
		return WrapError(err)
	}
	bytes := msg.Unwrap().Raw()

	// Send message
	return s.ch.Send(ctx, bytes)
}

// SendEnd sends an end message to the channel.
func (ch *serverChannel) SendEnd(ctx async.Context) status.Status {
	s, ok := ch.acquire()
	if !ok {
		return status.Closed
	}
	defer ch.release()

	// Lock send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	if s.sendEnd {
		return Error("end already sent")
	}

	// Make message
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	msg, err := s.sendBuilder.buildEnd(buf)
	if err != nil {
		return WrapError(err)
	}
	bytes := msg.Unwrap().Raw()

	// Send message
	s.sendEnd = true
	return s.ch.Send(ctx, bytes)
}

// SendResponse sends a response and closes the channel.
func (ch *serverChannel) SendResponse(ctx async.Context, result []byte, st status.Status) status.Status {
	s, ok := ch.acquire()
	if !ok {
		return status.Closed
	}
	defer ch.release()

	// Lock send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	// Make message
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	msg, err := s.sendBuilder.buildResponse(buf, result, st)
	if err != nil {
		return WrapError(err)
	}
	bytes := msg.Unwrap().Raw()

	// Send message
	return s.ch.SendAndClose(ctx, bytes)
}

// Internal

// Free frees the channel.
func (ch *serverChannel) Free() {
	if ch.freed.Swap(true) {
		return
	}

	ch.release()
}

// private

func (ch *serverChannel) acquire() (*serverChannelState, bool) {
	acquired := ch.refs.Acquire()
	if acquired {
		s := ch.state.Load()
		return s, true
	}

	ch.release()
	return nil, false
}

func (ch *serverChannel) release() {
	released := ch.refs.Release()
	if !released {
		return
	}

	ch.free()
}

func (ch *serverChannel) free() {
	s := ch.state.Swap(nil)
	if s == nil {
		return
	}

	releaseServerState(s)
}

// state

var serverStatePool = pools.NewPoolFunc(newServerChannelState)

func acquireServerState() *serverChannelState {
	return serverStatePool.New()
}

func releaseServerState(s *serverChannelState) {
	s.reset()
	serverStatePool.Put(s)
}

func (s *serverChannelState) reset() {
	s.ch = nil
	s.method = s.method[:0]

	s.sendReq = false
	s.sendEnd = false

	s.recvReq = prpc.Request{}
	s.recvEnd = false
	s.recvFailed = false
	s.recvError = status.None
}

func (s *serverChannelState) readFail(st status.Status) {
	s.recvFailed = true
	s.recvError = st
}
