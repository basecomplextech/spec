package rpc

import (
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/pools"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/mpx"
	"github.com/basecomplextech/spec/proto/prpc"
)

// Channel is a client RPC channel.
type Channel interface {
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

	// Response

	// Response receives a response and returns its status and result if status is OK.
	Response(ctx async.Context) (ref.R[spec.Value], status.Status)

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
	method string

	// send
	sendLock async.Lock
	sendReq  bool // request sent
	sendEnd  bool // end sent
	sendBuf  *alloc.Buffer
	sendMsg  spec.Writer

	// rev
	recvLock async.Lock
	recvEnd  bool // end received
	recvResp bool // response received

	recvFailed bool
	recvError  status.Status

	// temp result stores result until Response is called
	result   ref.R[spec.Value]
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
	sendBuf := alloc.NewBuffer()

	return &channelState{
		sendLock: async.NewLock(),
		sendBuf:  sendBuf,
		sendMsg:  spec.NewWriterBuffer(sendBuf),

		recvLock: async.NewLock(),
	}
}

// Request sends a request to the server.
func (ch *channel) Request(ctx async.Context, req prpc.Request) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Lock send
	select {
	case <-s.sendLock:
	case <-ctx.Wait():
		return ctx.Status()
	}
	defer s.sendLock.Unlock()

	if s.sendReq {
		return Error("request already sent")
	}

	// Make request
	var msg []byte
	{
		s.sendBuf.Reset()
		s.sendMsg.Reset(s.sendBuf)

		w := prpc.NewMessageWriterTo(s.sendMsg.Message())
		w.Type(prpc.MessageType_Request)
		w.CopyReq(req)

		p, err := w.Build()
		if err != nil {
			return WrapError(err)
		}

		msg = p.Unwrap().Raw()
	}

	// Send request
	s.method = requestMethod(req)
	s.sendReq = true
	return s.ch.Send(ctx, msg)
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
	select {
	case <-s.recvLock:
	case <-ctx.Wait():
		return nil, false, ctx.Status()
	}
	defer s.recvLock.Unlock()

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

// Send

// Send sends a message to the channel.
func (ch *channel) Send(ctx async.Context, message []byte) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Lock send
	select {
	case <-s.sendLock:
	case <-ctx.Wait():
		return ctx.Status()
	}
	defer s.sendLock.Unlock()

	if s.sendEnd {
		return Error("end already sent")
	}

	// Make message
	var msg []byte
	{
		s.sendBuf.Reset()
		s.sendMsg.Reset(s.sendBuf)

		w := prpc.NewMessageWriterTo(s.sendMsg.Message())
		w.Type(prpc.MessageType_Message)
		w.Msg(message)

		p, err := w.Build()
		if err != nil {
			return WrapError(err)
		}

		msg = p.Unwrap().Raw()
	}

	// Send message
	return s.ch.Send(ctx, msg)
}

// SendEnd sends an end message to the channel.
func (ch *channel) SendEnd(ctx async.Context) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Lock send
	select {
	case <-s.sendLock:
	case <-ctx.Wait():
		return ctx.Status()
	}
	defer s.sendLock.Unlock()

	if s.sendEnd {
		return Error("end already sent")
	}

	// Make message
	var msg []byte
	{
		s.sendBuf.Reset()
		s.sendMsg.Reset(s.sendBuf)

		w := prpc.NewMessageWriterTo(s.sendMsg.Message())
		w.Type(prpc.MessageType_End)

		p, err := w.Build()
		if err != nil {
			return WrapError(err)
		}

		msg = p.Unwrap().Raw()
	}

	// Send message
	s.sendEnd = true
	return s.ch.Send(ctx, msg)
}

// Response

// Response receives a response and returns its status and result if status is OK.
func (ch *channel) Response(ctx async.Context) (ref.R[spec.Value], status.Status) {
	s, ok := ch.rlock()
	if !ok {
		return nil, status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Lock receive
	select {
	case <-s.recvLock:
	case <-ctx.Wait():
		return nil, ctx.Status()
	}
	defer s.recvLock.Unlock()

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

	// Read messages
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
	select {
	case s.sendLock <- struct{}{}:
	default:
	}
	select {
	case s.recvLock <- struct{}{}:
	default:
	}

	s.ch = nil
	s.logger = nil
	s.method = ""

	s.sendReq = false
	s.sendEnd = false
	s.sendBuf.Reset()
	s.sendMsg.Reset(s.sendBuf)

	s.recvEnd = false
	s.recvResp = false
	s.recvFailed = false
	s.recvError = status.None

	if s.result != nil {
		s.result.Release()
		s.result = nil
		s.resultOK = false
		s.resultSt = status.None
	}
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

func parseResult(resp prpc.Response) (ref.R[spec.Value], status.Status) {
	// Parse status
	st := parseStatus(resp.Status())
	if !st.OK() {
		return nil, st
	}

	// Return nil when no result
	result := resp.Result()
	if len(result) == 0 {
		return ref.NewNoop[spec.Value](nil), status.OK
	}

	// Copy result to buffer
	buf := alloc.AcquireBuffer()
	done := false
	defer func() {
		if !done {
			buf.Free()
		}
	}()
	buf.Write(resp.Result())

	// Parse result
	v, err := spec.NewValueErr(buf.Bytes())
	if err != nil {
		return nil, WrapError(err)
	}

	// Wrap into ref
	ref := ref.NewFreer(v, buf)
	done = true
	return ref, status.OK
}
