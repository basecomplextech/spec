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
	// Request returns the request, the message is valid until the next call to Read.
	Request(cancel <-chan struct{}) (prpc.Request, status.Status)

	// Read

	// Read reads and returns a message, or false/end.
	// The method does not block if no messages, and returns false instead.
	// The message is valid until the next call to Read.
	Read(cancel <-chan struct{}) ([]byte, bool, status.Status)

	// ReadSync reads and returns a message from the channel, or an end.
	// The method blocks until a message is received, or the channel is closed.
	// The message is valid until the next call to Read.
	ReadSync(cancel <-chan struct{}) ([]byte, status.Status)

	// ReadWait returns a channel which is notified on a new message, or a channel close.
	ReadWait() <-chan struct{}

	// Write

	// Write writes a message to the channel.
	Write(cancel <-chan struct{}, message []byte) status.Status

	// WriteEnd writes an end message to the channel.
	WriteEnd(cancel <-chan struct{}) status.Status
}

// internal

var _ ServerChannel = (*serverChannel)(nil)

type serverChannel struct {
	ch tcp.Channel

	stateMu sync.RWMutex
	state   *serverChannelState
}

func newServerChannel(ch tcp.Channel, req prpc.Request) *serverChannel {
	s := acquireServerState()
	s.readReq = req

	return &serverChannel{
		ch:    ch,
		state: s,
	}
}

// Request returns the request, the message is valid until the next call to ReadSync.
func (ch *serverChannel) Request(cancel <-chan struct{}) (prpc.Request, status.Status) {
	s, ok := ch.rlock()
	if !ok {
		return prpc.Request{}, status.Closed
	}
	defer ch.stateMu.RUnlock()

	// Read lock
	select {
	case <-s.readLock:
	case <-cancel:
		return prpc.Request{}, status.Cancelled
	}
	defer s.readLock.Unlock()

	// Check state
	if s.readReq.IsEmpty() {
		return prpc.Request{}, Error("request already received")
	}

	return s.readReq, status.OK
}

// Read

// Read reads and returns a message, or false/end.
// The method does not block if no messages, and returns false instead.
// The message is valid until the next call to Read.
func (ch *serverChannel) Read(cancel <-chan struct{}) ([]byte, bool, status.Status) {
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

	// Clear request
	if !s.readReq.IsEmpty() {
		s.readReq = prpc.Request{}
	}

	// Read message
	var msg prpc.Message
	{
		b, ok, st := ch.ch.Read(cancel)
		switch {
		case !st.OK():
			return nil, false, st
		case !ok:
			return nil, false, status.OK
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
		s.readEnd = true
		return nil, false, status.End
	}

	st := Errorf("unexpected message type %d", typ)
	s.readFail(st)
	return nil, false, st
}

// ReadSync reads and returns a message from the channel, or an end.
// The method blocks until a message is received, or the channel is closed.
// The message is valid until the next call to Read.
func (ch *serverChannel) ReadSync(cancel <-chan struct{}) ([]byte, status.Status) {
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
		case <-ch.ch.Wait():
		}
	}
}

// ReadWait returns a channel which is notified on a new message, or a channel close.
func (ch *serverChannel) ReadWait() <-chan struct{} {
	return ch.ch.Wait()
}

// Write

// Write writes a message to the channel.
func (ch *serverChannel) Write(cancel <-chan struct{}, message []byte) status.Status {
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

// WriteEnd writes an end message to the channel.
func (ch *serverChannel) WriteEnd(cancel <-chan struct{}) status.Status {
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
	readReq    prpc.Request
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

	s.readReq = prpc.Request{}
	s.readEnd = false
	s.readFailed = false
	s.readError = status.None
}

func (s *serverChannelState) readFail(st status.Status) {
	s.readFailed = true
	s.readError = st
}
