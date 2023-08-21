package tcp

import (
	"sync"

	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/status"
)

// Stream is a single stream in a TCP connection.
type Stream interface {
	// Status returns the status of the stream.
	Status() status.Status

	// Methods

	// Read reads a message from the stream, the message is valid until the next iteration.
	Read(cancel <-chan struct{}) ([]byte, status.Status)

	// Write writes a message to the stream.
	Write(msg []byte) status.Status

	// Internal

	// Free closes the stream and releases its resources.
	Free()
}

// internal

var _ Stream = (*stream)(nil)

type stream struct {
	conn *conn

	id bin.Bin128

	mu sync.Mutex
	st status.Status

	readBuf    *messageBuffer
	readChan   chan struct{}
	readWaiter bool

	// enforce single reader/writer
	readMu  sync.Mutex
	writeMu sync.Mutex
}

func newStream(conn *conn, id bin.Bin128) *stream {
	return &stream{
		conn: conn,

		id: id,
		st: status.OK,

		readBuf:  newMessageBuffer(),
		readChan: make(chan struct{}, 1),
	}
}

// Status returns the status of the stream.
func (s *stream) Status() status.Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.st
}

// Methods

// Read reads a message from the stream, the message is valid until the next iteration.
func (s *stream) Read(cancel <-chan struct{}) ([]byte, status.Status) {
	// Enforce single reader
	s.readMu.Lock()
	defer s.readMu.Unlock()

	for {
		// Try to read next message
		msg, ok, st := s.read()
		switch {
		case !st.OK():
			return nil, st
		case ok:
			return msg, status.OK
		}

		// Wait for next message
		select {
		case <-cancel:
			return nil, status.Cancelled
		case <-s.readChan:
			continue
		}
	}
}

// Write writes a message to the stream.
func (s *stream) Write(msg []byte) status.Status {
	// Enforce single writer
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	// Check status
	s.mu.Lock()
	if !s.st.OK() {
		st := s.st
		s.mu.Unlock()
		return st
	}
	s.mu.Unlock()

	// Write message
	ok, st := s.conn.writeStream(s.id, msg)
	switch {
	case !st.OK():
		return st
	case ok:
		return status.OK
	}

	// Return error if failed
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.st
}

// Internal

// Free closes the stream and releases its resources.
func (s *stream) Free() {
	s.close()
	s.conn.closeStream(s.id)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.readBuf.free()
}

// internal

// close closes the stream.
func (s *stream) close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.st.OK() {
		return
	}

	// Close stream
	s.st = status.End
	if !s.readWaiter {
		return
	}

	// Notify waiter
	select {
	case s.readChan <- struct{}{}:
	default:
	}
}

func (s *stream) read() ([]byte, bool, status.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read next message
	msg, ok := s.readBuf.next()
	if ok {
		s.readWaiter = false
		return msg, true, status.OK
	}

	// Clear channel
	select {
	case <-s.readChan:
	default:
	}

	// Add waiter, this is a small performance optimization
	s.readWaiter = true
	return nil, false, s.st
}

// receive receives a message from a connection.
func (s *stream) receive(msg []byte) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.st.OK() {
		return false
	}

	// Append message
	s.readBuf.append(msg)
	if !s.readWaiter {
		return true
	}

	// Notify waiter
	select {
	case s.readChan <- struct{}{}:
	default:
	}
	return true
}
