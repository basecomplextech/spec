package tcp

import (
	"net"
	"sync"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/ptcp"
)

// Conn is a TCP network connection.
type Conn interface {
	// Accept accepts a new stream, or returns an end status.
	Accept(cancel <-chan struct{}) (Stream, status.Status)

	// Open opens a new stream and immediately writes a message.
	Open(cancel <-chan struct{}) (Stream, status.Status)

	// Internal

	// Free closes and frees the connection.
	Free()
}

// Dial dials an address and returns a connection.
func Dial(address string, logger logging.Logger) (Conn, status.Status) {
	nc, err := net.Dial("tcp", address)
	if err != nil {
		return nil, tcpError(err)
	}

	c := newConn(nc, true /* client */, logger)
	c.Run()
	return c, status.OK
}

// internal

var _ Conn = (*conn)(nil)

type conn struct {
	conn   net.Conn
	client bool
	logger logging.Logger

	accept *acceptQueue
	reader *reader
	writer *writer

	mu      sync.RWMutex
	st      status.Status
	main    async.Routine[struct{}]
	streams map[bin.Bin128]*stream

	pendingMu   sync.Mutex
	pending     *idset
	pendingChan chan struct{}
}

func newConn(c net.Conn, client bool, logger logging.Logger) *conn {
	return &conn{
		conn:   c,
		client: client,

		accept: newAcceptQueue(),
		reader: newReader(c, client),
		writer: newWriter(c, client),

		st:      status.OK,
		streams: make(map[bin.Bin128]*stream),

		pending:     newIDSet(),
		pendingChan: make(chan struct{}, 1),
	}
}

// Accept accepts a new stream, or returns an end status.
func (c *conn) Accept(cancel <-chan struct{}) (Stream, status.Status) {
	c.mu.Lock()
	if !c.st.OK() {
		c.mu.Unlock()
		return nil, c.st
	}
	c.mu.Unlock()

	for {
		stream, ok, st := c.accept.poll()
		switch {
		case !st.OK():
			return nil, st
		case ok:
			return stream, status.OK
		}

		select {
		case <-cancel:
			return nil, status.Cancelled
		case <-c.accept.wait():
			continue
		}
	}
}

// Open opens a new stream and immediately writes a message.
func (c *conn) Open(cancel <-chan struct{}) (Stream, status.Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.st.OK() {
		return nil, c.st
	}

	// Get unique id
	var id bin.Bin128
	for i := 0; i < 10; i++ {
		id = bin.Random128()
		if _, ok := c.streams[id]; !ok {
			break
		}
		id = bin.Bin128{}
	}
	if id == (bin.Bin128{}) {
		return nil, tcpErrorf("failed to generate unique stream id")
	}

	// Create stream
	s := newStream(id, c)
	c.streams[id] = s

	// Write message
	st := c.writer.writeOpenStream(id)
	if !st.OK() {
		s.closeBoth()
		delete(c.streams, id)
		if debug {
			debugPrint(c.client, "open failed\t", id, st)
		}
		return nil, st
	}
	return s, status.OK
}

// Free closes and frees the connection.
func (c *conn) Free() {
	c.close()
}

// Run starts the connection main goroutine.
func (c *conn) Run() async.Routine[struct{}] {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.main != nil {
		return c.main
	}

	c.main = async.Go(c.run)
	return c.main
}

// internal

func (c *conn) run(cancel <-chan struct{}) status.Status {
	defer func() {
		if e := recover(); e != nil {
			st, stack := status.RecoverStack(e)
			c.logger.Error("Internal connection panic", "status", st, "stack", string(stack))
		}
	}()

	// Start loops
	read := async.Go(c.readLoop)
	write := async.Go(c.writeLoop)
	defer async.CancelWaitAll(read, write)
	defer c.close()

	// Wait cancel/exit
	var st status.Status
	select {
	case <-cancel:
		st = status.Cancelled
	case <-read.Wait():
		st = read.Status()
	case <-write.Wait():
		st = write.Status()
	}

	// Check status
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeEnd,
		codeConnClosed:
		return st
	}

	// Log internal errors
	c.logger.Debug("Internal connection error", "status", st)
	return st
}

// close

func (c *conn) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.st.OK() {
		return
	}

	defer c._closeStreams()
	c.st = statusConnClosed
	c.conn.Close()
}

func (c *conn) _closeStreams() {
	for _, s := range c.streams {
		s.closeBoth()
	}
}

// read loop

func (c *conn) readLoop(cancel <-chan struct{}) status.Status {
	for {
		// Receive message
		msg, st := c.reader.read()
		if !st.OK() {
			return st
		}

		// Handle message
		code := msg.Code()
		switch code {
		case ptcp.Code_OpenStream:
			m := msg.Open()
			if st := c.handleNewStream(m); !st.OK() {
				return st
			}

		case ptcp.Code_CloseStream:
			m := msg.Close()
			if st := c.handleStreamClosed(m); !st.OK() {
				return st
			}

		case ptcp.Code_StreamMessage:
			m := msg.Message()
			if st := c.handleStreamMessage(m); !st.OK() {
				return st
			}

		default:
			return tcpErrorf("unexpected tpc message code %d", code)
		}
	}
}

// write loop

func (c *conn) writeLoop(cancel <-chan struct{}) status.Status {
	active := make(map[bin.Bin128]struct{})
	pending := newIDSet()

	for {
		// Handle active streams

		for {
			// Add pending streams
			pending = c.switchPending(pending)
			for _, id := range pending.list {
				active[id] = struct{}{}
			}

			// Write stream messages
			for id := range active {
				s, ok, st := c.getStream(id)
				switch {
				case !st.OK():
					return st
				case !ok:
					delete(active, id)
					continue
				}

				// Pull message from write queue
				msg, ok, st := s.pull()
				switch {
				case !st.OK():
					// Queue end, close stream, write close message
					delete(active, id)
					c.closeLocal(id)

					if st := c.writer.writeCloseStream(id); !st.OK() {
						return st
					}

				case !ok:
					// No mesages, remove from active
					delete(active, id)

				case ok:
					// Write data message
					if st := c.writer.writeStreamMessage(id, msg); !st.OK() {
						return st
					}
				}
			}

			// No more streams, break
			if len(active) == 0 {
				break
			}
		}

		// Flush buffered writes before waiting
		if st := c.writer.flush(); !st.OK() {
			return st
		}

		// Wait for active streams
		select {
		case <-cancel:
			return status.Cancelled
		case <-c.pendingChan:
		}
	}
}

// streams

// getStream returns a stream by id.
func (c *conn) getStream(id bin.Bin128) (*stream, bool, status.Status) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.st.OK() {
		return nil, false, c.st
	}

	s, ok := c.streams[id]
	if !ok {
		return nil, false, status.OK
	}
	return s, true, status.OK
}

// closeLocal closes a local stream when all outgoing messages are sent and write queue is ended.
func (c *conn) closeLocal(id bin.Bin128) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.st.OK() {
		return
	}

	// Ignore absent
	s, ok := c.streams[id]
	if !ok {
		return
	}

	// Close both queues inside connection, and free the stream.
	// When local queue is closed, it means that the local stream
	// is out of scope, and has been freed. No need to keep it anymore.
	//
	// Also it prevents from leaking streams when the remote side
	// fails to close the stream.
	s.closeBoth()
	delete(c.streams, id)
}

// receive

// handleNewStream handles an open stream message, or writes an error on duplicate.
func (c *conn) handleNewStream(msg ptcp.OpenStream) status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.st.OK() {
		return c.st
	}

	// Check duplicate
	id := msg.Id()
	if _, ok := c.streams[id]; ok {
		return c.writer.writeCloseStream(id)
	}

	// Create stream
	s := newStream(id, c)
	c.streams[id] = s

	c.accept.push(s)
	return status.OK
}

// handleStreamClosed handles a close stream message.
func (c *conn) handleStreamClosed(msg ptcp.CloseStream) status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.st.OK() {
		return c.st
	}

	// Ignore absent
	id := msg.Id()
	s, ok := c.streams[id]
	if !ok {
		return status.OK
	}

	// Close remote queue
	s.closeRemote()

	// Delete stream if both queues closed
	if s.closed() {
		delete(c.streams, id)
	}
	return status.OK
}

// handleStreamMessage handles a stream message.
func (c *conn) handleStreamMessage(msg ptcp.StreamMessage) status.Status {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.st.OK() {
		return c.st
	}

	id := msg.Id()
	s, ok := c.streams[id]
	if !ok {
		if debug {
			debugPrint(c.client, "no stream\t", id)
		}
		return status.OK
	}

	data := msg.Data()
	_, _ = s.push(data) // ignore status, stream may be closed
	return status.OK
}

// notify

// notify adds a stream to the pending list.
func (c *conn) notify(id bin.Bin128) {
	c.pendingMu.Lock()
	defer c.pendingMu.Unlock()

	ok := c.pending.add(id)
	if !ok {
		return
	}

	if c.pending.len() > 1 {
		return
	}

	select {
	case c.pendingChan <- struct{}{}:
	default:
	}
}

// switchPending switches the pending idset.
func (c *conn) switchPending(next *idset) *idset {
	next.clear()

	c.pendingMu.Lock()
	defer c.pendingMu.Unlock()

	if c.pending.len() == 0 {
		return next
	}

	prev := c.pending
	c.pending = next
	return prev
}
