package tcp

import (
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/status"
)

// Stream is a single stream in a TCP connection.
type Stream interface {
	// Read reads a message from the stream, the message is valid until the next iteration.
	Read(cancel <-chan struct{}) ([]byte, status.Status)

	// Write writes a message to the stream.
	Write(cancel <-chan struct{}, msg []byte) status.Status

	// Internal

	// Free closes the stream and releases its resources.
	Free()
}

// internal

var _ Stream = (*stream)(nil)

type stream struct {
	id     bin.Bin128
	conn   *conn
	client bool

	mu      sync.Mutex
	closed  bool // queues closed for writing
	removed bool // removed from the connection
	freed   bool // queues freed

	remoteOpened bool
	remoteClosed bool
	freeCalled   bool

	queues *streamQueues
}

func openStream(id bin.Bin128, conn *conn, requestNil []byte) *stream {
	queues := acquireQueues()

	if debug {
		debugPrint(conn.client, "stream.open\t", id)
	}

	s := &stream{
		id:     id,
		conn:   conn,
		client: conn.client,

		queues: queues,
	}

	if len(requestNil) > 0 {
		s.queues.out.Write(requestNil)
	}
	return s
}

func openedStream1(id bin.Bin128, conn *conn, request []byte) *stream {
	queues := acquireQueues()

	if debug {
		debugPrint(conn.client, "stream.opened\t", id)
	}

	s := &stream{
		id:   id,
		conn: conn,

		remoteOpened: true,
		queues:       queues,
	}
	if len(request) > 0 {
		s.queues.in.Write(request)
	}
	return s
}

// Read reads a message from the stream, the message is valid until the next iteration.
func (s *stream) Read(cancel <-chan struct{}) ([]byte, status.Status) {
	for {
		queue, st := s.in()
		if !st.OK() {
			return nil, st
		}

		// Try to read next message
		msg, ok, st := queue.Read()
		if debug && ok {
			debugPrint(s.client, "stream.read\t", s.id, ok, st)
		}

		switch {
		case !st.OK():
			return nil, st
		case ok:
			return msg, status.OK
		}

		// Wait for next message
		if debug {
			debugPrint(s.client, "stream.rwait\t", s.id)
		}

		select {
		case <-cancel:
			return nil, status.Cancelled
		case <-queue.Wait():
		}
	}
}

// Write writes a message to the stream.
func (s *stream) Write(cancel <-chan struct{}, msg []byte) status.Status {
	for {
		out, st := s.out()
		if !st.OK() {
			return st
		}

		// Try to write message
		ok, wasEmpty, st := out.Write(msg)
		if debug && ok {
			debugPrint(s.client, "stream.write\t", s.id, ok, st)
		}

		switch {
		case !st.OK():
			return st
		case ok:
			if wasEmpty {
				s.conn.enqueue(s.id)
			}
			return status.OK
		}

		// Wait for more space
		if debug {
			debugPrint(s.client, "stream.wwait\t", s.id)
		}

		select {
		case <-cancel:
			return status.Cancelled
		case <-out.WaitCanWrite(len(msg)):
		}
	}
}

// Internal

// Free closes the stream and releases its resources.
func (s *stream) Free() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.freeCalled = true

	if s.closed {
		s.maybeRemove()
		return
	}

	if debug {
		debugPrint(s.client, "stream.close\t", s.id)
	}

	s.closed = true
	s.queues.close()

	s.conn.enqueue(s.id)
	s.maybeRemove()
}

// internal

func (s *stream) connClosed() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.removed = true

	if s.closed {
		s.maybeFree()
		return
	}

	s.closed = true
	s.remoteClosed = true

	s.queues.out.Clear()
	s.queues.closeWithError(statusClosed)

	s.maybeFree()
}

func (s *stream) receiveClosed() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.remoteClosed {
		return
	}

	if debug {
		debugPrint(s.client, "stream.received-closed\t", s.id)
	}

	s.closed = true
	s.remoteClosed = true

	s.queues.out.Clear()
	s.queues.close()

	s.maybeRemove()
}

func (s *stream) receiveMessage(msg []byte) {
	// in, st := s.in()
	// if !st.OK() {
	// 	return
	// }

	ok, _, st := s.queues.in.Write(msg)
	if !ok || !st.OK() {
		return
	}

	if debug {
		debugPrint(s.client, "stream.received-message\t", s.id)
	}

	// s.mu.Lock()
	// defer s.mu.Unlock()

	// if s.closed {
	// 	return
	// }
}

func (s *stream) write(w *writer) (bool, status.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Free if removed
	if s.removed {
		s.maybeFree()
		return false, status.OK
	}

	// Send open
	open := !s.closed
	if open && !s.remoteOpened {
		msg, _, st := s.queues.out.Read()
		if !st.OK() {
			return false, st
		}

		s.remoteOpened = true
		if st := w.writeOpenStream(s.id, msg); !st.OK() {
			return false, st
		}
	}

	// Send message, ignore status,
	// we use closed flag instead
	msg, ok, st1 := s.queues.out.Read()
	if ok {
		if st := w.writeStreamMessage(s.id, msg); !st.OK() {
			return false, st
		}

		// Return if not end, otherwise continue to send close
		if st1.OK() {
			return true, status.OK
		}
	}

	// Send close
	if s.closed && !s.remoteClosed {
		s.remoteClosed = true

		if st := w.writeCloseStream(s.id); !st.OK() {
			return false, st
		}
		s.maybeRemove()
	}

	return false, status.OK
}

// private

func (s *stream) in() (alloc.BufferQueue, status.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.freed {
		return nil, tcpErrorf("stream freed")
	}
	return s.queues.in, status.OK
}

func (s *stream) out() (alloc.BufferQueue, status.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.freed {
		return nil, tcpErrorf("stream freed")
	}
	return s.queues.out, status.OK
}

func (s *stream) maybeRemove() {
	if s.removed {
		s.maybeFree()
		return
	}

	canRemove := s.closed && s.remoteClosed
	if !canRemove {
		return
	}

	if debug {
		debugPrint(s.client, "stream.removed\t", s.id)
	}

	s.removed = true
	s.conn.remove(s.id)
	s.maybeFree()
}

func (s *stream) maybeFree() {
	if s.freed {
		return
	}

	canFree := s.removed && s.freeCalled
	if !canFree {
		return
	}

	if debug {
		debugPrint(s.client, "stream.freed\t", s.id)
	}

	q := s.queues
	s.freed = true
	s.queues = closedQueues
	releaseQueues(q)
}

// queues

var queuePool = &sync.Pool{}
var closedQueues = func() *streamQueues {
	q := newStreamQueues()
	q.close()
	return q
}()

type streamQueues struct {
	in  alloc.BufferQueue
	out alloc.BufferQueue
}

func newStreamQueues() *streamQueues {
	return &streamQueues{
		in:  alloc.NewBufferQueue(), // unbounded
		out: alloc.NewBufferQueueCap(streamWriteQueueCap),
	}
}

func acquireQueues() *streamQueues {
	obj := queuePool.Get()
	if obj == nil {
		return newStreamQueues()
	}
	return obj.(*streamQueues)
}

func releaseQueues(q *streamQueues) {
	q.reset()
	queuePool.Put(q)
}

func (q *streamQueues) close() {
	q.in.Close()
	q.out.Close()
}

func (q *streamQueues) closeWithError(st status.Status) {
	q.in.CloseWithError(st)
	q.out.CloseWithError(st)
}

func (q *streamQueues) reset() {
	q.in.Reset()
	q.out.Reset()
}
