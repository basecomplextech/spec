package rpc

import (
	"sync"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/status"
)

var _ Transport = (*TestTransport)(nil)

type TestTransport struct {
	mu sync.Mutex

	Stream *TestStream
}

func NewTestTransport() *TestTransport {
	return &TestTransport{
		Stream: NewTestStream(),
	}
}

// Free releases the transport.
func (t *TestTransport) Free() {}

// Open opens a stream.
func (t *TestTransport) Open(cancel <-chan struct{}) (TransportStream, status.Status) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Stream == nil || t.Stream.closed {
		t.Stream = NewTestStream()
	}
	return t.Stream, status.OK
}

// Push pushes an incoming message.
func (t *TestTransport) Push(msg []byte) {
	t.Stream.Push(msg)
}

// Pop pops an outgoing message.
func (t *TestTransport) Pop() ([]byte, bool) {
	return t.Stream.Pop()
}

// Stream

var _ TransportStream = (*TestStream)(nil)

type TestStream struct {
	mu         sync.RWMutex
	closed     bool
	closedChan chan struct{}

	incoming async.Queue[[]byte]
	outgoing async.Queue[[]byte]
}

func NewTestStream() *TestStream {
	return &TestStream{
		closedChan: make(chan struct{}),
		incoming:   async.NewQueue[[]byte](),
		outgoing:   async.NewQueue[[]byte](),
	}
}

// Free releases the stream.
func (s *TestStream) Free() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	close(s.closedChan)

	s.incoming.Close()
	s.outgoing.Close()
}

// Send sends a message.
func (s *TestStream) Send(cancel <-chan struct{}, msg []byte) status.Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return status.Errorf("stream closed")
	}

	s.outgoing.Push(msg)
	return status.OK
}

// Receive receives a message, the message is valid until the next call to Receive.
func (s *TestStream) Receive(cancel <-chan struct{}) ([]byte, status.Status) {
loop:
	for {
		// Try to pop incoming message
		msg, ok := s.incoming.Pop()
		if ok {
			return msg, status.OK
		}

		// Wait for incoming message
		select {
		case <-cancel:
			return nil, status.Cancelled
		case <-s.incoming.Wait():
			continue loop
		case <-s.closedChan:
			return nil, status.End
		}
	}
}

// Push pushes an incoming message.
func (s *TestStream) Push(msg []byte) {
	s.incoming.Push(msg)
}

// Pop pops an outgoing message.
func (s *TestStream) Pop() ([]byte, bool) {
	return s.outgoing.Pop()
}
