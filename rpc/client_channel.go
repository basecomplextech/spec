package rpc

import (
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/prpc"
	"github.com/basecomplextech/spec/tcp"
)

// Channel is a client RPC channel.
type Channel interface {
	// Read

	// Receive receives a message from the channel, or an end.
	// The method blocks until a message is received, or the channel is closed.
	// The message is valid until the next call to Read/Receive/Response.
	Receive(cancel <-chan struct{}) ([]byte, status.Status)

	// Read reads and returns a message, or false.
	// The method does not block if no messages, and returns false instead.
	// The message is valid until the next call to Read/Receive/Response.
	Read(cancel <-chan struct{}) ([]byte, bool, status.Status)

	// Wait returns a channel which is notified on a new message, or a channel close.
	Wait() <-chan struct{}

	// Write

	// Write writes a message to the channel.
	Write(cancel <-chan struct{}, message []byte) status.Status

	// End writes an end message to the channel.
	End(cancel <-chan struct{}) status.Status

	// Response

	// Response receives a response and returns its status and result if status is OK.
	Response(cancel <-chan struct{}) (*ref.R[spec.Value], status.Status)

	// Internal

	// Free frees the channel.
	Free()
}

// internal

var _ Channel = (*channel)(nil)

type channel struct {
	ch tcp.Channel

	stateMu sync.RWMutex
	state   *channelState
}

func newChannel(ch tcp.Channel) *channel {
	s := acquireState()

	return &channel{
		ch:    ch,
		state: s,
	}
}

// Request sends a request to the server.
func (ch *channel) Request(cancel <-chan struct{}, req prpc.Request) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Write lock
	select {
	case <-s.writeLock:
	case <-cancel:
		return status.Cancelled
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
	s.writeReq = true
	return ch.ch.Write(cancel, msg)
}

// Read

// Receive receives a message from the channel, or an end.
// The method blocks until a message is received, or the channel is closed.
// The message is valid until the next call to Read/Receive/Response.
func (ch *channel) Receive(cancel <-chan struct{}) ([]byte, status.Status) {
	for {
		msg, ok, st := ch.Read(cancel)
		switch {
		case !st.OK():
			return nil, st
		case ok:
			return msg, status.OK
		}

		select {
		case <-cancel:
			return nil, status.Cancelled
		case <-ch.Wait():
		}
	}
}

// Read reads and returns a message, or false.
// The method does not block if no messages, and returns false instead.
// The message is valid until the next call to Read/Receive/Response.
func (ch *channel) Read(cancel <-chan struct{}) ([]byte, bool, status.Status) {
	s, ok := ch.rlock()
	if !ok {
		return nil, false, status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Read lock
	select {
	case <-s.readLock:
	case <-cancel:
		return nil, false, status.Cancelled
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
	msg, ok, st := ch.read(cancel)
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

// Wait returns a channel which is notified on a new message, or a channel close.
func (ch *channel) Wait() <-chan struct{} {
	return ch.ch.Wait()
}

// Write

// Write writes a message to the channel.
func (ch *channel) Write(cancel <-chan struct{}, message []byte) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Write lock
	select {
	case <-s.writeLock:
	case <-cancel:
		return status.Cancelled
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
	return ch.ch.Write(cancel, msg)
}

// End writes an end message to the channel.
func (ch *channel) End(cancel <-chan struct{}) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Write lock
	select {
	case <-s.writeLock:
	case <-cancel:
		return status.Cancelled
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
	return ch.ch.Write(cancel, msg)
}

// Response

// Response receives a response and returns its status and result if status is OK.
func (ch *channel) Response(cancel <-chan struct{}) (*ref.R[spec.Value], status.Status) {
	s, ok := ch.rlock()
	if !ok {
		return nil, status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Read lock
	select {
	case <-s.readLock:
	case <-cancel:
		return nil, status.Cancelled
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
		msg, st := ch.receive(cancel)
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
func (ch *channel) read(cancel <-chan struct{}) (prpc.Message, bool, status.Status) {
	b, ok, st := ch.ch.Read(cancel)
	switch {
	case !st.OK():
		return prpc.Message{}, false, st
	case !ok:
		return prpc.Message{}, false, status.OK
	}

	msg, _, err := prpc.ParseMessage(b)
	if err != nil {
		return prpc.Message{}, false, WrapError(err)
	}
	return msg, true, status.OK
}

// receive receives, parses and returns the next message, or blocks.
func (ch *channel) receive(cancel <-chan struct{}) (prpc.Message, status.Status) {
	b, st := ch.ch.Receive(cancel)
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

var statePool = &sync.Pool{}

type channelState struct {
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
	result   *ref.R[spec.Value]
	resultOK bool
	resultSt status.Status
}

func acquireState() *channelState {
	v := statePool.Get()
	if v != nil {
		return v.(*channelState)
	}

	return newChannelState()
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
	s.readFailed = true
	s.readError = st
}

// util

func parseResult(resp prpc.Response) (*ref.R[spec.Value], status.Status) {
	// Parse status
	st := parseStatus(resp.Status())
	if !st.OK() {
		return nil, st
	}

	// Return nil when no result
	result := resp.Result()
	if len(result) == 0 {
		return ref.NewNoFreer[spec.Value](nil), status.OK
	}

	// Copy result to buffer
	buf := alloc.NewBuffer()
	buf.Write(resp.Result())

	ok := false
	defer func() {
		if !ok {
			buf.Free()
		}
	}()

	// Parse result
	v, err := spec.NewValueErr(buf.Bytes())
	if err != nil {
		return nil, WrapError(err)
	}

	// Wrap into ref
	ref := ref.NewFreer(v, buf)
	ok = true
	return ref, status.OK
}
