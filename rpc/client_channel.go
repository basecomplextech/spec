package rpc

import (
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/pools"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/mpx"
	"github.com/basecomplextech/spec/proto/prpc"
)

// Channel is a client RPC channel.
type Channel interface {
	// Context returns a channel context.
	Context() Context

	// Send

	// Send sends a message to the channel.
	Send(ctx async.Context, message []byte) status.Status

	// SendEnd sends an end message to the channel.
	SendEnd(ctx async.Context) status.Status

	// Receive

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

	// Response receives a response and returns its status and result if status is OK.
	//
	// The message is valid until the channel is freed.
	Response(ctx async.Context) (spec.Value, status.Status)

	// Internal

	// Free frees the channel.
	Free()
}

// internal

var _ Channel = (*channel)(nil)

type channel struct {
	stateMu sync.RWMutex
	state   *channelState
}

type channelState struct {
	ch     mpx.Channel
	logger logging.Logger
	method []byte

	// send
	sendMu      sync.Mutex
	sendReq     bool // request sent
	sendEnd     bool // end sent
	sendBuilder builder

	// rev
	recvMu   sync.Mutex
	recvEnd  bool // end received
	recvResp bool // response received

	recvFailed bool
	recvError  status.Status

	// temp result stores result until Response is called
	result   spec.Value
	resultOK bool
	resultSt status.Status
}

func newChannel(ch mpx.Channel, logger logging.Logger) *channel {
	s := acquireState()
	s.ch = ch
	s.logger = logger

	return &channel{state: s}
}

func newChannelState() *channelState {
	return &channelState{
		sendBuilder: newBuilder(),
	}
}

// Method returns a joined method name, the string is valid until the channel is freed.
func (ch *channel) Method() string {
	s, ok := ch.rlock()
	if !ok {
		return ""
	}
	defer ch.stateMu.RUnlock()

	return unsafeString(s.method)
}

// Context returns a channel context.
func (ch *channel) Context() Context {
	s, ok := ch.rlock()
	if !ok {
		return mpx.ClosedContext()
	}
	defer ch.stateMu.RUnlock()

	return s.ch.Context()
}

// Send

// Request sends a request to the server.
func (ch *channel) Request(ctx async.Context, req prpc.Request) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Lock send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	if s.sendReq {
		return Error("request already sent")
	}

	// Make request
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	msg, err := s.sendBuilder.buildRequest(buf, req)
	if err != nil {
		return WrapError(err)
	}
	bytes := msg.Unwrap().Raw()

	// Send request
	s.method = requestMethod(s.method, req)
	s.sendReq = true
	return s.ch.Send(ctx, bytes)
}

// Send sends a message to the channel.
func (ch *channel) Send(ctx async.Context, message []byte) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return status.Closed
	}
	defer ch.stateMu.RUnlock()

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
func (ch *channel) SendEnd(ctx async.Context) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return status.Closed
	}
	defer ch.stateMu.RUnlock()

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

// Receive

// Receive receives and returns a message from the channel, or an end.
//
// The method blocks until a message is received, or the channel is closed.
// The message is valid until the next call to Receive/ReceiveAsync.
func (ch *channel) Receive(ctx async.Context) ([]byte, status.Status) {
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
func (ch *channel) ReceiveAsync(ctx async.Context) ([]byte, bool, status.Status) {
	s, ok := ch.rlock()
	if !ok {
		return nil, false, status.Closed
	}
	defer ch.stateMu.RUnlock()

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

	// Read message
	msg, ok, st := s.receiveAsync(ctx)
	switch {
	case !st.OK():
		s.receiveFail(st)
		return nil, false, st
	case !ok:
		return nil, false, status.OK
	}

	// Handle message
	typ := msg.Type()
	switch typ {
	case prpc.MessageType_Message:
		return msg.Msg(), true, status.OK

	case prpc.MessageType_End:
		s.recvEnd = true
		return nil, false, status.End

	case prpc.MessageType_Response:
		s.recvEnd = true
		s.recvResp = true

		result, st := parseResult(msg.Resp())
		s.result = result
		s.resultOK = true
		s.resultSt = st
		return nil, false, status.End
	}

	st = Errorf("unexpected message type %d", typ)
	s.receiveFail(st)
	return nil, false, st
}

// ReceiveWait returns a channel which is notified on a new message, or a channel close.
func (ch *channel) ReceiveWait() <-chan struct{} {
	s, ok := ch.rlock()
	if !ok {
		return closedChan
	}
	defer ch.stateMu.RUnlock()

	return s.ch.ReceiveWait()
}

// Response receives a response and returns its status and result if status is OK.
//
// The message is valid until the channel is freed.
func (ch *channel) Response(ctx async.Context) (spec.Value, status.Status) {
	s, ok := ch.rlock()
	if !ok {
		return nil, status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Lock receive
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	// Check state
	switch {
	case s.recvFailed:
		return nil, s.recvError

	case s.recvResp:
		if !s.resultOK {
			return nil, Error("response already received")
		}

		st := s.resultSt
		result := s.result

		s.result = nil
		s.resultOK = false
		return result, st

	}

	// Receive messages
	for {
		msg, st := s.receive(ctx)
		if !st.OK() {
			s.receiveFail(st)
			return nil, st
		}

		// Handle message
		typ := msg.Type()
		switch typ {
		case prpc.MessageType_Message:
			continue // Skip messages until response

		case prpc.MessageType_End:
			s.recvEnd = true
			continue // Skip messages until response

		case prpc.MessageType_Response:
			s.recvResp = true
			return parseResult(msg.Resp())
		}

		st = Errorf("unexpected message type %d", typ)
		s.receiveFail(st)
		return nil, st
	}
}

// Internal

// Free frees the channel.
func (ch *channel) Free() {
	ch.stateMu.Lock()
	defer ch.stateMu.Unlock()

	s := ch.state
	if s == nil {
		return
	}

	defer s.ch.Free()
	ch.state = nil
	releaseState(s)
}

// private

func (ch *channel) rlock() (*channelState, bool) {
	ch.stateMu.RLock()

	if ch.state == nil {
		ch.stateMu.RUnlock()
		return nil, false
	}

	return ch.state, true
}

// receive receives, parses and returns the next message, or blocks.
func (s *channelState) receive(ctx async.Context) (prpc.Message, status.Status) {
	b, st := s.ch.Receive(ctx)
	if !st.OK() {
		return prpc.Message{}, st
	}

	msg, _, err := prpc.ParseMessage(b)
	if err != nil {
		return prpc.Message{}, WrapError(err)
	}
	return msg, status.OK
}

// receiveAsync receives, parses and returns the next message, or false.
func (s *channelState) receiveAsync(ctx async.Context) (prpc.Message, bool, status.Status) {
	b, ok, st := s.ch.ReceiveAsync(ctx)
	if !ok || !st.OK() {
		return prpc.Message{}, false, st
	}

	msg, _, err := prpc.ParseMessage(b)
	if err != nil {
		return prpc.Message{}, false, WrapError(err)
	}
	return msg, true, status.OK
}

// state

var statePool = pools.MakePool(newChannelState)

func acquireState() *channelState {
	return statePool.New()
}

func releaseState(s *channelState) {
	s.reset()
	statePool.Put(s)
}

func (s *channelState) reset() {
	s.ch = nil
	s.logger = nil
	s.method = s.method[:0]

	s.sendReq = false
	s.sendEnd = false

	s.recvEnd = false
	s.recvResp = false
	s.recvFailed = false
	s.recvError = status.None

	s.result = nil
	s.resultOK = false
	s.resultSt = status.None
}

func (s *channelState) receiveFail(st status.Status) {
	if s.recvFailed {
		return
	}

	s.recvFailed = true
	s.recvError = st

	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeExternalError,
		status.CodeEnd,
		status.CodeWait:

	case status.CodeError, status.CodeCorrupted:
		s.logger.ErrorStatus("RPC client request error", st, "method", s.method)

	default:
		if s.logger.DebugEnabled() {
			s.logger.Debug("RPC client request error", "status", st, "method", s.method)
		}
	}
}

// util

func parseResult(resp prpc.Response) (spec.Value, status.Status) {
	// Parse status
	st := parseStatus(resp.Status())
	if !st.OK() {
		return nil, st
	}

	// Parse result
	result := resp.Result()
	if len(result) == 0 {
		return nil, status.OK
	}
	return result, status.OK
}
