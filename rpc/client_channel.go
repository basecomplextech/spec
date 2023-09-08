package rpc

import (
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/prpc"
	"github.com/basecomplextech/spec/tcp"
)

// Channel is a client RPC channel.
type Channel interface {
	// End sends an end message to the channel.
	End(cancel <-chan struct{}) status.Status

	// Send sends a message to the channel.
	Send(cancel <-chan struct{}, message []byte) status.Status

	// Receive receives a message from the channel, or an end.
	// The message is valid until the next call to Receive or Response.
	Receive(cancel <-chan struct{}) ([]byte, status.Status)

	// Response receives a response and returns its status and result if status is OK.
	Response(cancel <-chan struct{}) (*alloc.Buffer, status.Status)

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
		ch: ch,

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

// End sends an end message to the channel.
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

// Send sends a message to the channel.
func (ch *channel) Send(cancel <-chan struct{}, message []byte) status.Status {
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

// Receive receives a message from the channel, or an end.
// The message is valid until the next call to Receive or Response.
func (ch *channel) Receive(cancel <-chan struct{}) ([]byte, status.Status) {
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
	case s.readEnd:
		return nil, status.End
	}

	// Read message
	msg, st := ch.read(cancel)
	if !st.OK() {
		s.readFail(st)
		return nil, st
	}

	// Handle message
	typ := msg.Type()
	switch typ {
	case prpc.MessageType_Message:
		return msg.Msg(), status.OK

	case prpc.MessageType_End:
		s.readEnd = true
		return nil, status.End

	case prpc.MessageType_Response:
		s.readEnd = true
		s.readResp = true

		result, st := parseResult(msg.Resp())
		s.result = result
		s.resultOK = true
		s.resultSt = st
		return nil, status.End
	}

	st = Errorf("unexpected message type %d", typ)
	s.readFail(st)
	return nil, st
}

// Response receives a response and returns its status and result if status is OK.
func (ch *channel) Response(cancel <-chan struct{}) (*alloc.Buffer, status.Status) {
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
		msg, st := ch.read(cancel)
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

// read reads, parses and returns the next message.
func (ch *channel) read(cancel <-chan struct{}) (prpc.Message, status.Status) {
	b, st := ch.ch.Read(cancel)
	if !st.OK() {
		return prpc.Message{}, st
	}

	msg, _, err := prpc.ParseMessage(b)
	if err != nil {
		return prpc.Message{}, WrapError(err)
	}
	return msg, status.OK
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
	result   *alloc.Buffer
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

	if s.result != nil {
		s.result.Free()
		s.result = nil
	}
}

func (s *channelState) readFail(st status.Status) {
	s.readFailed = true
	s.readError = st
}

// util

func parseResult(resp prpc.Response) (*alloc.Buffer, status.Status) {
	st := parseStatus(resp.Status())
	if !st.OK() {
		return nil, st
	}

	buf := alloc.NewBuffer()
	buf.Write(resp.Result())
	return buf, status.OK
}
