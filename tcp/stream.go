package tcp

import (
	"fmt"
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
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
	Write(cancel <-chan struct{}, msg []byte) status.Status

	// Internal

	// Free closes the stream and releases its resources.
	Free()
}

// internal

const writeQueueSize = 1 << 17 // 128kb

var _ Stream = (*stream)(nil)

type stream struct {
	id   bin.Bin128
	conn *conn

	mu sync.Mutex
	st status.Status

	readQueue  alloc.BufferQueue
	writeQueue alloc.BufferQueue
}

func newStream(id bin.Bin128, conn *conn) *stream {
	return &stream{
		id:   id,
		conn: conn,
		st:   status.OK,

		readQueue:  alloc.NewBufferQueue(), // unbounded
		writeQueue: alloc.NewBufferQueueCap(writeQueueSize),
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
	for {
		// Try to read next message
		msg, ok, st := s.readQueue.Read()
		if debug {
			fmt.Println("<- read", ok, st, msg)
		}
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
		case <-s.readQueue.Wait():
			continue
		}
	}
}

// Write writes a message to the stream.
func (s *stream) Write(cancel <-chan struct{}, msg []byte) status.Status {
	for {
		// Try to write message
		ok, st := s.writeQueue.Write(msg)
		if debug {
			fmt.Println("-> write", ok, st, msg)
		}
		switch {
		case !st.OK():
			return st
		case ok:
			s.notify()
			return status.OK
		}

		// Wait for more space
		select {
		case <-cancel:
			return status.Cancelled
		case <-s.writeQueue.WaitCanWrite(len(msg)):
		}
	}
}

// Internal

// Free closes the stream and releases its resources.
func (s *stream) Free() {
	defer s.free()

	s.close()
	s.notify()
}

// internal

func (s *stream) free() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.readQueue.Free()
	s.writeQueue.Free()
}

// close closes the stream.
func (s *stream) close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.st.OK() {
		return
	}

	s.st = statusStreamClosed
	s.readQueue.Close()
	s.writeQueue.Close()
}

// pull/push

// pull pulls an outgoing message from the stream write queue.
func (s *stream) pull() ([]byte, bool, status.Status) {
	return s.writeQueue.Read()
}

// push pushes an incoming message to the stream read queue.
func (s *stream) push(msg []byte) (bool, status.Status) {
	return s.readQueue.Write(msg)
}

func (s *stream) notify() {
	s.conn.notify(s.id)
}
