package tcp

import (
	"net"
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
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

	// Open opens a new stream.
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

type conn struct {
	conn   net.Conn
	client bool
	logger logging.Logger

	acceptQueue *acceptQueue
	writeQueue  alloc.BufferQueue

	reader *reader
	writer *writer

	mu      sync.RWMutex
	st      status.Status
	main    async.Routine[struct{}]
	streams map[bin.Bin128]*stream
}

func newConn(c net.Conn, client bool, logger logging.Logger) *conn {
	return &conn{
		conn:   c,
		client: client,

		acceptQueue: newAcceptQueue(),
		writeQueue:  alloc.NewBufferQueue(),

		reader: newReader(c, client),
		writer: newWriter(c, client),

		st:      status.OK,
		streams: make(map[bin.Bin128]*stream),
	}
}

// Accept accepts a new stream, or returns an end status.
func (c *conn) Accept(cancel <-chan struct{}) (Stream, status.Status) {
	for {
		s, ok, st := c.acceptQueue.poll()
		switch {
		case !st.OK():
			return nil, st
		case ok:
			return s, status.OK
		}

		select {
		case <-cancel:
			return nil, status.Cancelled
		case <-c.acceptQueue.wait():
			continue
		}
	}
}

// Open opens a new stream.
func (c *conn) Open(cancel <-chan struct{}) (Stream, status.Status) {
	return c.open()
}

// Internal

// Free closes and frees the connection.
func (c *conn) Free() {
	var main async.Routine[struct{}]

	c.mu.Lock()
	main = c.main
	c.mu.Unlock()

	if main == nil {
		c.close()
	} else {
		main.Cancel()
	}
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
	defer c.close()

	// Start loops
	read := async.Go(c.readLoop)
	write := async.Go(c.writeLoop)
	defer async.CancelWaitAll(read, write)
	defer c.conn.Close()

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
		codeClosed:
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
	defer c.closeStreams()

	c.st = statusClosed
	c.acceptQueue.close()
	c.writeQueue.Close()
	c.conn.Close()

	if debug {
		debugPrint(c.client, "conn.close\t", c.st)
	}
}

func (c *conn) closeStreams() {
	for _, s := range c.streams {
		s.connClosed()
	}
}

// readLoop

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
			if st := c.handleOpened(msg); !st.OK() {
				return st
			}

		case ptcp.Code_CloseStream:
			if st := c.handleClosed(msg); !st.OK() {
				return st
			}

		case ptcp.Code_StreamMessage:
			if st := c.handleMessage(msg); !st.OK() {
				return st
			}

		default:
			return tcpErrorf("unexpected tpc message code %d", code)
		}
	}
}

func (c *conn) handleOpened(msg ptcp.Message) status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	m := msg.Open()
	id := m.Id()
	_, ok := c.streams[id]
	if ok {
		return tcpErrorf("stream %v already exists", id) // impossible
	}

	if debug {
		debugPrint(c.client, "conn.opened\t", id)
	}

	s := openedStream(id, c)
	c.streams[id] = s
	c.acceptQueue.push(s)
	return status.OK
}

func (c *conn) handleClosed(msg ptcp.Message) status.Status {
	m := msg.Close()
	id := m.Id()

	s, ok := c.remove(id)
	if !ok {
		return status.OK
	}

	s.receive(msg)
	return status.OK
}

func (c *conn) handleMessage(msg ptcp.Message) status.Status {
	m := msg.Message()
	id := m.Id()

	s, ok := c.get(id)
	if !ok {
		return status.OK
	}

	s.receive(msg)
	return status.OK
}

// write loop

func (c *conn) writeLoop(cancel <-chan struct{}) status.Status {
	for {
		b, ok, st := c.writeQueue.Read()
		switch {
		case !st.OK():
			return st
		case ok:
			if st := c.writeMessage(b); !st.OK() {
				return st
			}
			continue
		}

		// Flush buffered writes
		if st := c.writer.flush(); !st.OK() {
			return st
		}

		// Wait for messages
		select {
		case <-cancel:
			return status.Cancelled
		case <-c.writeQueue.Wait():
		}
	}
}

// writeMessage writes a message to
func (c *conn) writeMessage(b []byte) status.Status {
	msg := ptcp.NewMessage(b)
	code := msg.Code()

	if code == ptcp.Code_CloseStream {
		id := msg.Close().Id()
		c.remove(id)
	}

	return c.writer.write(b)
}

// streams

func (c *conn) get(id bin.Bin128) (*stream, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	s, ok := c.streams[id]
	return s, ok
}

func (c *conn) open() (*stream, status.Status) {
	id := bin.Random128()

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.st.OK() {
		return nil, c.st
	}

	_, ok := c.streams[id]
	if ok {
		return nil, tcpErrorf("failed to generate unique stream id") // impossible
	}

	if debug {
		debugPrint(c.client, "conn.open\t", id)
	}

	s := openStream(id, c)
	c.streams[id] = s
	return s, status.OK
}

func (c *conn) remove(id bin.Bin128) (*stream, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	s, ok := c.streams[id]
	if !ok {
		return nil, false
	}

	if debug {
		debugPrint(c.client, "conn.remove\t", id)
	}

	delete(c.streams, id)
	return s, true
}

// write writes an outgoing message, or returns a connection closed error.
func (c *conn) write(msg ptcp.Message) status.Status {
	b := msg.Unwrap().Raw()
	_, _, st := c.writeQueue.Write(b)
	if !st.OK() {
		return statusClosed
	}
	return status.OK
}
