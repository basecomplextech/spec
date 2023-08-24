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

	c := newConn(nc, logger)
	c.Run()
	return c, status.OK
}

// internal

var _ Conn = (*conn)(nil)

type conn struct {
	conn   net.Conn
	logger logging.Logger

	accept *acceptQueue
	reader *reader
	writer *writer

	mu      sync.RWMutex
	st      status.Status
	main    async.Routine[struct{}]
	streams map[bin.Bin128]*stream

	pendingMu sync.Mutex
	pending   *idset
	// pending     []bin.Bin128
	// pendingMap  map[bin.Bin128]struct{}
	pendingChan chan struct{}
}

func newConn(c net.Conn, logger logging.Logger) *conn {
	return &conn{
		conn: c,

		accept: newAcceptQueue(),
		reader: newReader(c),
		writer: newWriter(c),

		st:      status.OK,
		streams: make(map[bin.Bin128]*stream),

		pending: newIDSet(),
		// pendingMap:  make(map[bin.Bin128]struct{}),
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
	if st.OK() {
		return s, status.OK
	}

	// Delete on error
	defer s.free()
	delete(c.streams, id)
	return nil, st
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

	c.st = statusConnClosed
	defer c.conn.Close()

	for _, s := range c.streams {
		s.close()
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
			if st := c.handleOpenStream(m); !st.OK() {
				return st
			}

		case ptcp.Code_CloseStream:
			m := msg.Close()
			if st := c.handleCloseStream(m); !st.OK() {
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

				msg, ok, st := s.pull()
				switch {
				case !st.OK():
					// It's an end status, close stream
					delete(active, id)
					if st := c.closeStream(id); !st.OK() {
						return st
					}

				case ok:
					// Write stream message
					if st := c.writer.writeStreamMessage(id, msg); !st.OK() {
						return st
					}

				default:
					// No more stream messages
					delete(active, id)
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

		// Wait for more messages
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

// closeStream closes, deletes a stream and writes a close message.
func (c *conn) closeStream(id bin.Bin128) status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.st.OK() {
		return c.st
	}

	s, ok := c.streams[id]
	if !ok {
		return status.OK
	}

	s.close()
	delete(c.streams, id)
	return c.writer.writeCloseStream(id)
}

// handle

// handleOpenStream handles an open stream message, or writes an error on duplicate.
func (c *conn) handleOpenStream(msg ptcp.OpenStream) status.Status {
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

// handleCloseStream handles a close stream message, ignores absent streams.
func (c *conn) handleCloseStream(msg ptcp.CloseStream) status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.st.OK() {
		return c.st
	}

	id := msg.Id()
	s, ok := c.streams[id]
	if !ok {
		return status.OK
	}

	delete(c.streams, id)
	s.close()
	return status.OK
}

// handleStreamMessage handles a stream message, ignores absent streams.
func (c *conn) handleStreamMessage(msg ptcp.StreamMessage) status.Status {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.st.OK() {
		return c.st
	}

	id := msg.Id()
	s, ok := c.streams[id]
	if !ok {
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
