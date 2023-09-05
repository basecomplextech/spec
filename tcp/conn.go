package tcp

import (
	"net"
	"sync"
	"time"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/ptcp"
)

// Conn is a TCP network connection.
type Conn interface {
	// Stream opens a new stream.
	Stream(cancel <-chan struct{}) (Stream, status.Status)

	// Internal

	// Free closes and frees the connection.
	Free()
}

// Connect dials an address and returns a connection.
func Connect(address string, logger logging.Logger) (Conn, status.Status) {
	return connect(address, logger)
}

// ConnectTimeout dials an address and returns a connection.
func ConnectTimeout(address string, logger logging.Logger, timeout time.Duration) (Conn, status.Status) {
	return connectTimeout(address, logger, timeout)
}

// internal

type conn struct {
	conn    net.Conn
	client  bool
	handler Handler
	logger  logging.Logger

	reader     *reader
	writer     *writer
	writeQueue alloc.MQueue

	mu      sync.RWMutex
	st      status.Status
	main    async.Routine[struct{}]
	streams map[bin.Bin128]*stream
}

func connect(address string, logger logging.Logger) (*conn, status.Status) {
	nc, err := net.Dial("tcp", address)
	if err != nil {
		return nil, tcpError(err)
	}

	h := HandleFunc(func(s Stream) status.Status {
		return s.Close()
	})

	c := newConn(nc, true /* client */, h, logger)
	c.Run()
	return c, status.OK
}

func connectTimeout(address string, logger logging.Logger, timeout time.Duration) (*conn, status.Status) {
	nc, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil, tcpError(err)
	}

	h := HandleFunc(func(s Stream) status.Status {
		return s.Close()
	})

	c := newConn(nc, true /* client */, h, logger)
	c.Run()
	return c, status.OK
}

func newConn(c net.Conn, client bool, handlerNil Handler, logger logging.Logger) *conn {
	return &conn{
		conn:    c,
		client:  client,
		handler: handlerNil,
		logger:  logger,

		reader:     newReader(c, client),
		writer:     newWriter(c, client),
		writeQueue: alloc.NewMQueueCap(connWriteQueueCap),

		st:      status.OK,
		streams: make(map[bin.Bin128]*stream),
	}
}

// Stream opens a new stream.
func (c *conn) Stream(cancel <-chan struct{}) (Stream, status.Status) {
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
		<-main.Wait()
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
			c.logger.Error("Connection panic", "status", st, "stack", string(stack))
		}
	}()
	defer c.close()
	defer c.reader.free()

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
		status.CodeClosed:
		return st
	}

	// Log internal errors
	c.logger.Debug("Connection error", "status", st)
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

func (c *conn) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return !c.st.OK()
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
		case ptcp.Code_NewStream:
			if st := c.handleNewStream(msg); !st.OK() {
				return st
			}

		case ptcp.Code_CloseStream:
			if st := c.handleCloseStream(cancel, msg); !st.OK() {
				return st
			}

		case ptcp.Code_StreamMessage:
			if st := c.handleStreamMessage(cancel, msg); !st.OK() {
				return st
			}

		default:
			return tcpErrorf("unexpected tpc message code %d", code)
		}
	}
}

func (c *conn) handleNewStream(msg ptcp.Message) status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	m := msg.New()
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
	return status.OK
}

func (c *conn) handleCloseStream(cancel <-chan struct{}, msg ptcp.Message) status.Status {
	m := msg.Close()
	id := m.Id()

	s, ok := c.remove(id)
	if !ok {
		return status.OK
	}

	return s.receive(cancel, msg)
}

func (c *conn) handleStreamMessage(cancel <-chan struct{}, msg ptcp.Message) status.Status {
	m := msg.Message()
	id := m.Id()

	s, ok := c.get(id)
	if !ok {
		return status.OK
	}

	if !s.started {
		s.started = true
		go c.handleStream(s)
	}

	return s.receive(cancel, msg)
}

func (c *conn) handleStream(s *stream) {
	// No need to use async.Go here, because we don't need the result,
	// cancellation, and recover panics manually.
	defer func() {
		if e := recover(); e != nil {
			st, stack := status.RecoverStack(e)
			c.logger.Error("Stream panic", "status", st, "stack", string(stack))
		}
	}()
	defer s.Free()

	// Handle stream
	st := c.handler.HandleStream(s)
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeEnd,
		status.CodeClosed:
		return
	}

	// Log errors
	c.logger.Error("Stream error", "status", st)
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
		case <-c.writeQueue.ReadWait():
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
func (c *conn) write(cancel <-chan struct{}, msg ptcp.Message) status.Status {
	b := msg.Unwrap().Raw()

	for {
		ok, st := c.writeQueue.Write(b)
		switch {
		case !st.OK():
			return statusClosed
		case ok:
			return status.OK
		}

		// Wait for space
		select {
		case <-cancel:
			return status.Cancelled
		case <-c.writeQueue.WriteWait(len(b)):
			continue
		}
	}
}
