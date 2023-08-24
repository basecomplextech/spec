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

	// Free closes the stream.
	Free()
}

// internal

var _ Stream = (*stream)(nil)

type stream struct {
	id     bin.Bin128
	conn   *conn
	client bool

	mu     sync.Mutex
	queues *streamQueues

	freed bool
}

func newStream(id bin.Bin128, conn *conn) *stream {
	queues := acquireQueues()

	return &stream{
		id:     id,
		conn:   conn,
		client: conn.client,

		queues: queues,
	}
}

// Free closes the stream.
func (s *stream) Free() {
	s.closeLocal()
	s.notify()
}

// Methods

// Read reads a message from the stream, the message is valid until the next iteration.
func (s *stream) Read(cancel <-chan struct{}) ([]byte, status.Status) {
	for {
		// Try to read next message
		msg, ok, st := s.queues.read.Read()
		if debug {
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
			debugPrint(s.client, "stream.read-wait\t", s.id)
		}

		select {
		case <-cancel:
			return nil, status.Cancelled
		case <-s.queues.read.Wait():
		}
	}
}

// Write writes a message to the stream.
func (s *stream) Write(cancel <-chan struct{}, msg []byte) status.Status {
	for {
		// Try to write message
		ok, wasEmpty, st := s.queues.write.Write(msg)
		if debug {
			debugPrint(s.client, "stream.write\t", s.id, ok, st)
		}

		switch {
		case !st.OK():
			return st
		case ok:
			if wasEmpty {
				s.notify()
			}
			return status.OK
		}

		// Wait for more space
		if debug {
			debugPrint(s.client, "stream.write-wait\t", s.id)
		}
		select {
		case <-cancel:
			return status.Cancelled
		case <-s.queues.write.WaitCanWrite(len(msg)):
		}
	}
}

// internal

// closed returns true if both queues are closed.
func (s *stream) closed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	ok0 := s.queues.read.Closed()
	ok1 := s.queues.write.Closed()
	return ok0 && ok1
}

// closeBoth closes both queues.
func (s *stream) closeBoth() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.queues.close()
	s._maybeFree()
}

// closeLocal closes the write queue.
func (s *stream) closeLocal() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if debug {
		debugPrint(s.client, "stream.local-close\t", s.id)
	}

	s.queues.write.Close()
	s._maybeFree()
}

// closeRemote closes the read queue.
func (s *stream) closeRemote() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if debug {
		debugPrint(s.client, "stream.remote-close\t", s.id)
	}

	s.queues.read.Close()
	s._maybeFree()
}

// _maybeFree frees the stream if both queues are closed.
func (s *stream) _maybeFree() {
	switch {
	case s.freed:
		return
	case !s.queues.read.Closed():
		return
	case !s.queues.write.Closed():
		return
	}

	if debug {
		debugPrint(s.client, "stream.free\t", s.id)
	}

	q := s.queues
	s.freed = true
	s.queues = closedQueues
	releaseQueues(q)
}

// notify

// notify adds the stream to the pending connection streams.
func (s *stream) notify() {
	s.conn.notify(s.id)
}

// pull/push

// pull pulls an outgoing message from the stream write queue.
func (s *stream) pull() ([]byte, bool, status.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.queues.write.Read()
}

// push pushes an incoming message to the stream read queue.
func (s *stream) push(msg []byte) (bool, status.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ok, _, st := s.queues.read.Write(msg)
	return ok, st
}

// queues

var (
	queuePool = &sync.Pool{}

	closedQueues = func() *streamQueues {
		q := newStreamQueues()
		q.close()
		return q
	}()
)

type streamQueues struct {
	read  alloc.BufferQueue
	write alloc.BufferQueue
}

func newStreamQueues() *streamQueues {
	return &streamQueues{
		read:  alloc.NewBufferQueue(), // unbounded
		write: alloc.NewBufferQueueCap(streamWriteQueueCap),
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
	q.read.Close()
	q.write.Close()
}

func (q *streamQueues) reset() {
	q.read.Reset()
	q.write.Reset()
}
