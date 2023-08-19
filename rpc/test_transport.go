package rpc

import (
	"sync"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/status"
)

var _ Connector = (*TestConnector)(nil)

type TestConnector struct {
	mu sync.Mutex

	Conn *TestClientConn
}

func NewTestConnector() *TestConnector {
	return &TestConnector{
		Conn: NewTestClientConn(),
	}
}

// Free releases the transport.
func (t *TestConnector) Free() {}

// Connect connects to a server.
func (t *TestConnector) Connect(cancel <-chan struct{}) (ClientConn, status.Status) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Conn == nil || t.Conn.closed {
		t.Conn = NewTestClientConn()
	}
	return t.Conn, status.OK
}

// Push pushes an incoming message.
func (t *TestConnector) Push(msg []byte) {
	t.Conn.Push(msg)
}

// Pop pops an outgoing message.
func (t *TestConnector) Pop() ([]byte, bool) {
	return t.Conn.Pop()
}

// Conn

var _ ClientConn = (*TestClientConn)(nil)

type TestClientConn struct {
	mu         sync.RWMutex
	closed     bool
	closedChan chan struct{}

	incoming async.Queue[[]byte]
	outgoing async.Queue[[]byte]
}

func NewTestClientConn() *TestClientConn {
	return &TestClientConn{
		closedChan: make(chan struct{}),
		incoming:   async.NewQueue[[]byte](),
		outgoing:   async.NewQueue[[]byte](),
	}
}

// Free releases the stream.
func (s *TestClientConn) Free() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	close(s.closedChan)

	s.incoming.Close()
	s.outgoing.Close()
}

// Send sends a message.
func (s *TestClientConn) Send(cancel <-chan struct{}, msg []byte) status.Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return status.Errorf("stream closed")
	}

	s.outgoing.Push(msg)
	return status.OK
}

// Receive receives a message, the message is valid until the next call to Receive.
func (s *TestClientConn) Receive(cancel <-chan struct{}) ([]byte, status.Status) {
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
func (s *TestClientConn) Push(msg []byte) {
	s.incoming.Push(msg)
}

// Pop pops an outgoing message.
func (s *TestClientConn) Pop() ([]byte, bool) {
	return s.outgoing.Pop()
}
