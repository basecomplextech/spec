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
	id   bin.Bin128
	conn *conn

	// mutexes are used for single reader/writer
	readQueue *queue
	readMu    sync.Mutex
	writeMu   sync.Mutex

	mu sync.Mutex
	st status.Status
}

func newStream(conn *conn, id bin.Bin128) *stream {
	return &stream{
		conn: conn,

		id:        id,
		st:        status.OK,
		readQueue: newQueue(),
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
	s.readMu.Lock()
	defer s.readMu.Unlock()

	for {
		// Try to read next message
		msg, ok := s.readQueue.next()
		switch {
		case ok:
			return msg, status.OK
		}

		// Wait for next message
		select {
		case <-cancel:
			return nil, status.Cancelled
		case <-s.readQueue.wait():
			continue
		}
	}
}

// Write writes a message to the stream.
func (s *stream) Write(msg []byte) status.Status {
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
	return s.conn.sendMessage(s.id, msg)
}

// Internal

// Free closes the stream and releases its resources.
func (s *stream) Free() {
	defer s.free()

	s.close()
	s.conn.closeStream(s.id)
}

// internal

func (s *stream) free() {
	s.readQueue.free()
	s.readQueue = nil
}

// close closes the stream.
func (s *stream) close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.st.OK() {
		return
	}

	s.st = statusStreamClosed
	s.readQueue.close()
}

// receive

// receiveError receives an error from the connection.
func (s *stream) receiveError(st status.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.st.OK() {
		return
	}

	s.st = st
	s.readQueue.closeWithError(st)
}

// receiveClose receives a close message from the connection.
func (s *stream) receiveClose() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.st.OK() {
		return
	}

	s.st = statusStreamClosed
	s.readQueue.close()
}

// receiveMessage receives a message from the connection.
func (s *stream) receiveMessage(msg []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.st.OK() {
		return
	}

	s.readQueue.append(msg)
}
