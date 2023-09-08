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

// ServerChannel is a server RPC channel.
type ServerChannel interface {
	// End sends an end message to the channel.
	End(cancel <-chan struct{}) status.Status

	// Send sends a message to the channel.
	Send(cancel <-chan struct{}, message []byte) status.Status

	// Receive receives a message from the channel, or an end.
	// The message is valid until the next call to Receive or Response.
	Receive(cancel <-chan struct{}) ([]byte, status.Status)
}

// internal

var _ ServerChannel = (*serverChannel)(nil)

type serverChannel struct {
	ch tcp.Channel

	stateMu sync.RWMutex
	state   *serverChannelState
}

func newServerChannel(ch tcp.Channel) *serverChannel {
	s := acquireServerState()

	return &serverChannel{
		ch:    ch,
		state: s,
	}
}

// End sends an end message to the channel.
func (ch *serverChannel) End(cancel <-chan struct{}) status.Status {
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
func (ch *serverChannel) Send(cancel <-chan struct{}, message []byte) status.Status {
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
func (ch *serverChannel) Receive(cancel <-chan struct{}) ([]byte, status.Status) {
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
	var msg prpc.Message
	{
		b, st := ch.ch.Read(cancel)
		if !st.OK() {
			return nil, st
		}

		var err error
		msg, _, err = prpc.ParseMessage(b)
		if err != nil {
			return nil, WrapError(err)
		}
	}

	// Handle message
	typ := msg.Type()
	switch typ {
	case prpc.MessageType_Message:
		return msg.Msg(), status.OK

	case prpc.MessageType_End:
		s.readEnd = true
		return nil, status.End
	}

	st := Errorf("unexpected message type %d", typ)
	s.readFail(st)
	return nil, st
}

// Internal

// Free
func (ch *serverChannel) Free() {
	ch.stateMu.Lock()
	defer ch.stateMu.Unlock()

	if ch.state == nil {
		return
	}

	s := ch.state
	ch.state = nil
	releaseServerState(s)
}

// private

func (ch *serverChannel) rlock() (*serverChannelState, bool) {
	ch.stateMu.RLock()

	if ch.state == nil {
		ch.stateMu.RUnlock()
		return nil, false
	}

	return ch.state, true
}

// state

var serverStatePool = &sync.Pool{}

type serverChannelState struct {
	writeLock async.Lock
	writeReq  bool // request sent
	writeEnd  bool // end sent
	writeBuf  *alloc.Buffer
	writeMsg  spec.Writer

	readLock   async.Lock
	readEnd    bool // end received
	readFailed bool
	readError  status.Status
}

func acquireServerState() *serverChannelState {
	v := serverStatePool.Get()
	if v != nil {
		return v.(*serverChannelState)
	}

	return newServerChannelState()
}

func releaseServerState(s *serverChannelState) {
	s.reset()
	serverStatePool.Put(s)
}

func newServerChannelState() *serverChannelState {
	writeBuf := alloc.NewBuffer()

	return &serverChannelState{
		writeLock: async.NewLock(),
		writeBuf:  writeBuf,
		writeMsg:  spec.NewWriterBuffer(writeBuf),

		readLock: async.NewLock(),
	}
}

func (s *serverChannelState) reset() {
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
	s.readFailed = false
	s.readError = status.None
}

func (s *serverChannelState) readFail(st status.Status) {
	s.readFailed = true
	s.readError = st
}
