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

type Conn interface {
	// Accept accepts a new stream, or returns an end status.
	Accept(cancel <-chan struct{}) (Stream, status.Status)

	// Open opens a new stream and immediately sends a message.
	Open(cancel <-chan struct{}, msg []byte) (Stream, status.Status)

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

	acceptQueue *acceptQueue
	reader      *reader
	writer      *writer
	writeQueue  *writeQueue

	mu   sync.RWMutex
	st   status.Status
	main async.Routine[struct{}]

	streams map[bin.Bin128]*stream
}

func newConn(c net.Conn, logger logging.Logger) *conn {
	return &conn{
		conn: c,

		acceptQueue: newAcceptQueue(),
		reader:      newReader(c),
		writer:      newWriter(c),
		writeQueue:  newWriteQueue(),

		st:      status.OK,
		streams: make(map[bin.Bin128]*stream),
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
		stream, ok, st := c.acceptQueue.poll()
		switch {
		case !st.OK():
			return nil, st
		case ok:
			return stream, status.OK
		}

		select {
		case <-cancel:
			return nil, status.Cancelled
		case <-c.acceptQueue.wait():
			continue
		}
	}
}

// Open opens a new stream and immediately sends a message.
func (c *conn) Open(cancel <-chan struct{}, msg []byte) (Stream, status.Status) {
	return c.openStream(msg)
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

	// Start
	read := async.Go(c.readLoop)
	write := async.Go(c.writeLoop)
	defer async.CancelWaitAll(read, write)
	defer c.close()

	// Wait
	var st status.Status
	select {
	case <-cancel:
		st = status.Cancelled
	case <-read.Wait():
		st = read.Status()
	case <-write.Wait():
		st = write.Status()
	}

	// Log errors
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeEnd,
		codeConnClosed,
		codeStreamClosed:
		return st
	}

	c.logger.Debug("Internal connection error", "status", st)
	return st
}

// close

func (c *conn) close() {
	ok, st := c._close()
	if !ok {
		return
	}

	c._closeStreams(st)
}

func (c *conn) _close() (bool, status.Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.st.OK() {
		return false, c.st
	}

	c.st = statusConnClosed
	c.conn.Close()
	return true, c.st
}

func (c *conn) _closeStreams(st status.Status) {
	for _, s := range c.streams {
		s.receiveError(st)
	}

	c.streams = nil
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
			return tcpErrorf("unexpected tpc message code %d", code)
		}
	}
}

// write loop

func (c *conn) writeLoop(cancel <-chan struct{}) status.Status {
	for {
		// Write message
		msg, ok, st := c.writeQueue.queue.next()
		switch {
		case !st.OK():
			return st
		case ok:
			if st := c.writer.write(msg); !st.OK() {
				return st
			}
			continue
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

// receiveOpenStream is called by the reader on an open message.
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
	c.acceptQueue.push(s)
	return s, status.OK
}

// receiveCloseStream

// receiveCloseStream is called by the reader to close a stream.
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

// receiveStreamMessage is called by the reader to handle a stream message.
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
