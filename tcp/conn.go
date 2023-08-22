package tcp

import (
	"net"
	"sync"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/ptcp"
)

type Conn interface {
	// Open opens a new stream and immediately sends a message.
	Open(cancel <-chan struct{}, msg []byte) (Stream, status.Status)
}

// internal

var _ Conn = (*conn)(nil)

type conn struct {
	conn net.Conn

	server  bool
	handler Handler // only for server connections

	reader     *reader
	writer     *writer
	writeQueue *writeQueue

	mu      sync.RWMutex
	st      status.Status
	streams map[bin.Bin128]*stream
}

func newConn(c net.Conn) *conn {
	return &conn{
		conn: c,

		reader:     newReader(c),
		writer:     newWriter(c),
		writeQueue: newWriteQueue(),

		st:      status.OK,
		streams: make(map[bin.Bin128]*stream),
	}
}

// newClientCon returns a client connection.
func newClientCon(address string) (*conn, status.Status) {
	c, err := net.Dial("tcp", address)
	if err != nil {
		return nil, tcpError(err)
	}

	conn := newConn(c)
	go conn.run(nil)
	return conn, status.OK
}

func newServerConn(c net.Conn, handler Handler) *conn {
	conn := newConn(c)
	conn.server = true
	conn.handler = handler
	return conn
}

// Open opens a new stream and immediately sends a message.
func (c *conn) Open(cancel <-chan struct{}, msg []byte) (Stream, status.Status) {
	return c.openStream(msg)
}

// internal

func (c *conn) run(cancel <-chan struct{}) (st status.Status) {
	defer c.close() // for panics

	handler := async.Go(c.readLoop)
	writer := async.Go(c.writeLoop)
	defer async.CancelWaitAll(handler, writer)

	select {
	case <-cancel:
		st = status.Cancelled

	case <-handler.Wait():
		st = handler.Status()

	case <-writer.Wait():
		st = writer.Status()
	}

	c.close()
	return st
}

// close

func (c *conn) close() {
	streams, ok, st := c._close()
	if !ok {
		return
	}

	for _, s := range streams {
		s.receiveError(st)
	}
}

func (c *conn) _close() (map[bin.Bin128]*stream, bool, status.Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.st.OK() {
		return nil, false, c.st
	}

	c.st = statusConnClosed
	c.conn.Close()

	streams := c.streams
	c.streams = nil
	return streams, true, c.st
}

// read loop

func (c *conn) readLoop(cancel <-chan struct{}) status.Status {
	defer c.conn.Close()

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
			id := m.Id()
			data := m.Data()
			if st := c.receiveOpenStream(id, data); !st.OK() {
				return st
			}

		case ptcp.Code_CloseStream:
			m := msg.Close()
			id := m.Id()
			if st := c.receiveCloseStream(id); !st.OK() {
				return st
			}

		case ptcp.Code_StreamMessage:
			m := msg.Message()
			id := m.Id()
			data := m.Data()
			if st := c.receiveStreamMessage(id, data); !st.OK() {
				return st
			}
		default:
		}
	}
}

// write loop

func (c *conn) writeLoop(cancel <-chan struct{}) status.Status {
	defer c.conn.Close()

	for {
		// Write message
		msg, ok := c.writeQueue.queue.next()
		if ok {
			if st := c.writer.write(msg); !st.OK() {
				return st
			}
		}

		// Flush if no more messages
		if st := c.writer.flush(); !st.OK() {
			return st
		}

		// Wait for more messages
		select {
		case <-cancel:
			return status.Cancelled
		case <-c.writeQueue.queue.wait():
			continue
		}
	}
}

// internal

// openStream opens a new stream and immediately sends a message.
func (c *conn) openStream(data []byte) (Stream, status.Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check conn
	if !c.st.OK() {
		return nil, c.st
	}

	// Create stream
	// TODO: Check duplicates
	id := bin.Random128()
	s := newStream(c, id)
	c.streams[id] = s

	// Write message
	st := c.writeQueue.writeOpenStream(id, data)
	if st.OK() {
		return s, status.OK
	}

	// Close on error
	defer s.free()
	delete(c.streams, id)
	return nil, st
}

// closeStream is called by a stream to signal a close.
func (c *conn) closeStream(id bin.Bin128) status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check conn
	if !c.st.OK() {
		return c.st
	}

	// Check stream
	_, ok := c.streams[id]
	if !ok {
		return status.OK
	}

	// Delete stream
	delete(c.streams, id)
	return c.writeQueue.writeCloseStream(id)
}

// sendMessage is called by a stream to send a message.
func (c *conn) sendMessage(id bin.Bin128, b []byte) status.Status {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check conn
	if !c.st.OK() {
		return c.st
	}

	// Check stream
	_, ok := c.streams[id]
	if !ok {
		return statusStreamClosed
	}

	// Write message
	return c.writeQueue.writeStreamMessage(id, b)
}

// receiveOpenStream

// receiveOpenStream is called by the handler on an open message.
func (c *conn) receiveOpenStream(id bin.Bin128, data []byte) status.Status {
	// Create stream
	s, st := c._receiveOpenStream(id, data)
	if !st.OK() {
		return st
	}

	// Maybe pass message
	if len(data) > 0 {
		s.receiveMessage(data)
	}
	return status.OK
}

func (c *conn) _receiveOpenStream(id bin.Bin128, data []byte) (*stream, status.Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check conn
	if !c.st.OK() {
		return nil, c.st
	}

	// Create stream
	// TODO: Check duplicates
	s := newStream(c, id)
	c.streams[id] = s

	go c.handler.Handle(s)
	return s, status.OK
}

// receiveCloseStream

// receiveCloseStream is called by the handler to close a stream.
func (c *conn) receiveCloseStream(id bin.Bin128) status.Status {
	// Delete stream
	s, ok := c._receiveCloseStream(id)
	if !ok {
		return status.OK
	}

	// Notify stream
	s.receiveClose()
	return status.OK
}

func (c *conn) _receiveCloseStream(id bin.Bin128) (*stream, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check conn
	if !c.st.OK() {
		return nil, false
	}

	// Check stream
	s, ok := c.streams[id]
	if !ok {
		return nil, false
	}

	// Delete stream
	delete(c.streams, id)
	return s, true
}

// receiveStreamMessage

// receiveStreamMessage is called by the handler to handle a stream message.
func (c *conn) receiveStreamMessage(id bin.Bin128, data []byte) status.Status {
	// Get stream
	s, ok := c._receiveStreamMessage(id)
	if !ok {
		return status.OK
	}

	// Pass message
	s.receiveMessage(data)
	return status.OK
}

func (c *conn) _receiveStreamMessage(id bin.Bin128) (*stream, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.st.OK() {
		return nil, false
	}

	s, ok := c.streams[id]
	return s, ok
}
