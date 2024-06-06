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
	// Read reads and returns a message, or false/end.
	// The method does not block if no messages, and returns false instead.
	// The message is valid until the next call to Read/Response.
	Read(ctx async.Context) ([]byte, bool, status.Status)

	// ReadSync reads and returns a message from the channel, or an end.
	// The method blocks until a message is received, or the channel is closed.
	// The message is valid until the next call to Read/Response.
	ReadSync(ctx async.Context) ([]byte, status.Status)

	// ReadWait returns a channel which is notified on a new message, or a channel close.
	ReadWait() <-chan struct{}

	// Write

	// Write writes a message to the channel.
	Write(ctx async.Context, message []byte) status.Status

	// WriteEnd writes an end message to the channel.
	WriteEnd(ctx async.Context) status.Status

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
	ch mpx.Channel

	stateMu sync.RWMutex
	state   *channelState
}

func newChannel(ch mpx.Channel, logger logging.Logger) *channel {
	s := acquireState()
	s.logger = logger

	return &channel{
		ch:    ch,
		state: s,
	}
}

// Request sends a request to the server.
func (ch *channel) Request(ctx async.Context, req prpc.Request) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Write lock
	select {
	case <-s.writeLock:
	case <-ctx.Wait():
		return ctx.Status()
	}
	defer s.writeLock.Unlock()

	if s.writeReq {
		return Error("request already sent")
	}

	// Make request
	var msg []byte
	{
		s.writeBuf.Reset()
		s.writeMsg.Reset(s.writeBuf)

		w := prpc.NewMessageWriterTo(s.writeMsg.Message())
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
	s.writeReq = true
	return ch.ch.Send(ctx, msg)
}

// Read

// Read reads and returns a message, or false/end.
// The method does not block if no messages, and returns false instead.
// The message is valid until the next call to Read/Receive/Response.
func (ch *channel) Read(ctx async.Context) ([]byte, bool, status.Status) {
	s, ok := ch.rlock()
	if !ok {
		return nil, false, status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Read lock
	select {
	case <-s.readLock:
	case <-ctx.Wait():
		return nil, false, ctx.Status()
	}
	defer s.readLock.Unlock()

	// Check state
	switch {
	case s.readFailed:
		return nil, false, s.readError
	case s.readEnd:
		return nil, false, status.End
	}

	// Read message
	msg, ok, st := ch.read(ctx)
	switch {
	case !st.OK():
		s.readFail(st)
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
		s.readEnd = true
		return nil, false, status.End

	case prpc.MessageType_Response:
		s.readEnd = true
		s.readResp = true

		result, st := parseResult(msg.Resp())
		s.result = result
		s.resultOK = true
		s.resultSt = st
		return nil, false, status.End
	}

	st = Errorf("unexpected message type %d", typ)
	s.readFail(st)
	return nil, false, st
}

// ReadSync reads and returns a message from the channel, or an end.
// The method blocks until a message is received, or the channel is closed.
// The message is valid until the next call to Read/Receive/Response.
func (ch *channel) ReadSync(ctx async.Context) ([]byte, status.Status) {
	for {
		msg, ok, st := ch.Read(ctx)
		switch {
		case !st.OK():
			return nil, st
		case ok:
			return msg, status.OK
		}

		select {
		case <-ctx.Wait():
			return nil, ctx.Status()
		case <-ch.ch.ReceiveWait():
		}
	}
}

// ReadWait returns a channel which is notified on a new message, or a channel close.
func (ch *channel) ReadWait() <-chan struct{} {
	return ch.ch.ReceiveWait()
}

// Write

// Write writes a message to the channel.
func (ch *channel) Write(ctx async.Context, message []byte) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Write lock
	select {
	case <-s.writeLock:
	case <-ctx.Wait():
		return ctx.Status()
	}
	defer s.writeLock.Unlock()

	if s.writeEnd {
		return Error("end already sent")
	}

	// Make message
	var msg []byte
	{
		s.writeBuf.Reset()
		s.writeMsg.Reset(s.writeBuf)

		w := prpc.NewMessageWriterTo(s.writeMsg.Message())
		w.Type(prpc.MessageType_Message)
		w.Msg(message)

		p, err := w.Build()
		if err != nil {
			return WrapError(err)
		}

		msg = p.Unwrap().Raw()
	}

	// Send message
	return ch.ch.Send(ctx, msg)
}

// WriteEnd writes an end message to the channel.
func (ch *channel) WriteEnd(ctx async.Context) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Write lock
	select {
	case <-s.writeLock:
	case <-ctx.Wait():
		return ctx.Status()
	}
	defer s.writeLock.Unlock()

	if s.writeEnd {
		return Error("end already sent")
	}

	// Make message
	var msg []byte
	{
		s.writeBuf.Reset()
		s.writeMsg.Reset(s.writeBuf)

		w := prpc.NewMessageWriterTo(s.writeMsg.Message())
		w.Type(prpc.MessageType_End)

		p, err := w.Build()
		if err != nil {
			return WrapError(err)
		}

		msg = p.Unwrap().Raw()
	}

	// Send message
	s.writeEnd = true
	return ch.ch.Send(ctx, msg)
}

// Response

// Response receives a response and returns its status and result if status is OK.
func (ch *channel) Response(ctx async.Context) (ref.R[spec.Value], status.Status) {
	s, ok := ch.rlock()
	if !ok {
		return nil, status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Read lock
	select {
	case <-s.readLock:
	case <-ctx.Wait():
		return nil, ctx.Status()
	}
	defer s.readLock.Unlock()

	// Check state
	switch {
	case s.readFailed:
		return nil, s.readError

	case s.readResp:
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
		msg, st := ch.readSync(ctx)
		if !st.OK() {
			s.readFail(st)
			return nil, st
		}

		// Handle message
		typ := msg.Type()
		switch typ {
		case prpc.MessageType_Message:
			continue // Skip messages until response

		case prpc.MessageType_End:
			s.readEnd = true
			continue // Skip messages until response

		case prpc.MessageType_Response:
			s.readResp = true
			return parseResult(msg.Resp())
		}

		st = Errorf("unexpected message type %d", typ)
		s.readFail(st)
		return nil, st
	}
}

// Internal

// Free frees the channel.
func (ch *channel) Free() {
	ch.stateMu.Lock()
	defer ch.stateMu.Unlock()

	if ch.state == nil {
		return
	}
	defer ch.ch.Free()

	s := ch.state
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

// read reads, parses and returns the next message, or false.
func (ch *channel) read(ctx async.Context) (prpc.Message, bool, status.Status) {
	b, ok, st := ch.ch.ReceiveAsync(ctx)
	if !ok || !st.OK() {
		return prpc.Message{}, false, st
	}

	msg, _, err := prpc.ParseMessage(b)
	if err != nil {
		return prpc.Message{}, false, WrapError(err)
	}
	return msg, true, status.OK
}

// readSync reads, parses and returns the next message, or blocks.
func (ch *channel) readSync(ctx async.Context) (prpc.Message, status.Status) {
	b, st := ch.ch.Receive(ctx)
	if !st.OK() {
		return prpc.Message{}, st
	}

	msg, _, err := prpc.ParseMessage(b)
	if err != nil {
		return prpc.Message{}, WrapError(err)
	}
	return msg, status.OK
}

// state

var statePool = pools.MakePool(newChannelState)

type channelState struct {
	logger logging.Logger
	method string

	writeLock async.Lock
	writeReq  bool // request sent
	writeEnd  bool // end sent
	writeBuf  *alloc.Buffer
	writeMsg  spec.Writer

	readLock async.Lock
	readEnd  bool // end received
	readResp bool // response received

	readFailed bool
	readError  status.Status

	// temp result stores result until Response is called
	result   ref.R[spec.Value]
	resultOK bool
	resultSt status.Status
}

func acquireState() *channelState {
	return statePool.New()
}

func releaseState(s *channelState) {
	s.reset()
	statePool.Put(s)
}

func newChannelState() *channelState {
	writeBuf := alloc.NewBuffer()

	return &channelState{
		writeLock: async.NewLock(),
		writeBuf:  writeBuf,
		writeMsg:  spec.NewWriterBuffer(writeBuf),

		readLock: async.NewLock(),
	}
}

func (s *channelState) reset() {
	select {
	case s.writeLock <- struct{}{}:
	default:
	}
	select {
	case s.readLock <- struct{}{}:
	default:
	}

	s.logger = nil
	s.method = ""

	s.writeReq = false
	s.writeEnd = false
	s.writeBuf.Reset()
	s.writeMsg.Reset(s.writeBuf)

	s.readEnd = false
	s.readResp = false
	s.readFailed = false
	s.readError = status.None

	if s.result != nil {
		s.result.Release()
		s.result = nil
		s.resultOK = false
		s.resultSt = status.None
	}
}

func (s *channelState) readFail(st status.Status) {
	if s.readFailed {
		return
	}

	s.readFailed = true
	s.readError = st

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
	buf := acquireBuffer()
	buf.Write(resp.Result())
	free := func() {
		releaseBuffer(buf)
	}

	ok := false
	defer func() {
		if !ok {
			free()
		}
	}()

	// Parse result
	v, err := spec.NewValueErr(buf.Bytes())
	if err != nil {
		return nil, WrapError(err)
	}

	// Wrap into ref
	ref := ref.NewFree(v, free)
	ok = true
	return ref, status.OK
}
