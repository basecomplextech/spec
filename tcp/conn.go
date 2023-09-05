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
	// Close closes the connection.
	Close() status.Status

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
	handler Handler
	logger  logging.Logger

	client  bool
	socket  connSocket
	streams connStreams

	reader     *reader
	writer     *writer
	writeQueue alloc.MQueue

	routine async.Routine[struct{}]
}

func connect(address string, logger logging.Logger) (*conn, status.Status) {
	return connectTimeout(address, logger, 0)
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
	c.routine = async.Go(c.run)
	return c, status.OK
}

func newConn(c net.Conn, client bool, handler Handler, logger logging.Logger) *conn {
	return &conn{
		handler: handler,
		logger:  logger,

		client:  client,
		socket:  newConnSocket(client, c),
		streams: newConnStreams(client),

		reader:     newReader(c, client),
		writer:     newWriter(c, client),
		writeQueue: alloc.NewMQueueCap(connWriteQueueCap),
	}
}

// Close closes the connection.
func (c *conn) Close() status.Status {
	c.close()

	c.routine.Cancel()
	<-c.routine.Wait()
	return status.OK
}

// Stream opens a new stream.
func (c *conn) Stream(cancel <-chan struct{}) (Stream, status.Status) {
	return c.streams.open(c)
}

// Internal

// Free closes and frees the connection.
func (c *conn) Free() {
	defer c.reader.free()
	defer c.writeQueue.Free()

	c.Close()
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

	// Start loops
	reader := async.Go(c.readLoop)
	writer := async.Go(c.writeLoop)
	defer async.CancelWaitAll(reader, writer)
	defer c.close()

	// Wait cancel/exit
	var st status.Status
	select {
	case <-cancel:
		st = status.Cancelled
	case <-reader.Wait():
		st = reader.Status()
	case <-writer.Wait():
		st = writer.Status()
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
	c.logger.Error("Connection error", "client", c.client, "status", st)
	return st
}

// close

func (c *conn) close() {
	defer c.streams.close()
	defer c.writeQueue.Close()

	c.socket.close()
}

func (c *conn) closed() bool {
	return c.socket.closed()
}

// read

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
			m := msg.New()
			id := m.Id()

			if st := c.streams.opened(c, id); !st.OK() {
				return st
			}

		case ptcp.Code_CloseStream:
			m := msg.Close()
			id := m.Id()

			s, ok := c.streams.remove(id)
			if !ok {
				continue
			}

			if st := s.receive(cancel, msg); !st.OK() {
				return st
			}

		case ptcp.Code_StreamMessage:
			m := msg.Message()
			id := m.Id()

			s, ok := c.streams.get(id)
			if !ok {
				continue
			}

			if st := s.receive(cancel, msg); !st.OK() {
				return st
			}

		default:
			return tcpErrorf("unexpected tpc message code %d", code)
		}
	}
}

// write

func (c *conn) writeLoop(cancel <-chan struct{}) status.Status {
	for {
		b, ok, st := c.writeQueue.Read()
		switch {
		case !st.OK():
			return st

		case ok:
			msg := ptcp.NewMessage(b)
			code := msg.Code()

			if code == ptcp.Code_CloseStream {
				id := msg.Close().Id()
				c.streams.remove(id)
			}

			if st := c.writer.write(b); !st.OK() {
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

// write pushes an outgoing message to the write queue, or returns a connection closed error.
func (c *conn) write(cancel <-chan struct{}, msg ptcp.Message) status.Status {
	b := msg.Unwrap().Raw()

	for {
		ok, st := c.writeQueue.Write(b)
		switch {
		case !st.OK():
			return statusConnClosed
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

// socket

type connSocket struct {
	mu     sync.Mutex
	st     status.Status
	conn   net.Conn
	client bool
}

func newConnSocket(client bool, conn net.Conn) connSocket {
	return connSocket{
		st:     status.OK,
		conn:   conn,
		client: client,
	}
}

func (s *connSocket) close() status.Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.st.OK() {
		return status.OK
	}

	s.st = statusConnClosed
	s.conn.Close()

	if debug {
		debugPrint(s.client, "conn.close\t", s.st)
	}
	return s.st
}

func (s *connSocket) closed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return !s.st.OK()
}

func (s *connSocket) status() status.Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.st
}

// streams

type connStreams struct {
	mu      sync.Mutex
	client  bool
	closed  bool
	streams map[bin.Bin128]*stream
}

func newConnStreams(client bool) connStreams {
	return connStreams{
		client:  client,
		streams: make(map[bin.Bin128]*stream),
	}
}

func (s *connStreams) close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}
	s.closed = true

	for _, s := range s.streams {
		s.connClosed()
	}
}

func (s *connStreams) get(id bin.Bin128) (*stream, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, false
	}

	st, ok := s.streams[id]
	return st, ok
}

func (s *connStreams) open(c *conn) (*stream, status.Status) {
	id := bin.Random128()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, statusConnClosed
	}

	if debug {
		debugPrint(c.client, "conn.open\t", id)
	}

	stream := openStream(id, c)
	s.streams[stream.id] = stream
	return stream, status.OK
}

func (s *connStreams) opened(c *conn, id bin.Bin128) status.Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return statusConnClosed
	}

	_, ok := s.streams[id]
	if ok {
		return tcpErrorf("stream %v already exists", id) // impossible
	}

	if debug {
		debugPrint(c.client, "conn.opened\t", id)
	}

	stream := openedStream(id, c)
	s.streams[id] = stream
	return status.OK
}

func (s *connStreams) remove(id bin.Bin128) (*stream, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, false
	}

	st, ok := s.streams[id]
	if !ok {
		return nil, false
	}

	delete(s.streams, id)

	if debug {
		debugPrint(s.client, "conn.remove\t", id)
	}
	return st, true
}
