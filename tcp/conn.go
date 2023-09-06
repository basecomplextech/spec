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

	// Channel opens a new ch.
	Channel(cancel <-chan struct{}) (Channel, status.Status)

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

	client   bool
	socket   connSocket
	channels connChannels

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

	h := HandleFunc(func(c Channel) status.Status {
		return c.Close()
	})

	c := newConn(nc, true /* client */, h, logger)
	c.routine = async.Go(c.run)
	return c, status.OK
}

func newConn(c net.Conn, client bool, handler Handler, logger logging.Logger) *conn {
	return &conn{
		handler: handler,
		logger:  logger,

		client:   client,
		socket:   newConnSocket(client, c),
		channels: newConnChannels(client),

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

// Channel opens a new ch.
func (c *conn) Channel(cancel <-chan struct{}) (Channel, status.Status) {
	return c.channels.open(c)
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
	defer c.channels.close()
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
		case ptcp.Code_NewChannel:
			m := msg.New()
			id := m.Id()

			if st := c.channels.opened(c, id); !st.OK() {
				return st
			}

		case ptcp.Code_CloseChannel:
			m := msg.Close()
			id := m.Id()

			c, ok := c.channels.remove(id)
			if !ok {
				continue
			}

			if st := c.receive(cancel, msg); !st.OK() {
				return st
			}

		case ptcp.Code_ChannelMessage:
			m := msg.Message()
			id := m.Id()

			c, ok := c.channels.get(id)
			if !ok {
				continue
			}

			if st := c.receive(cancel, msg); !st.OK() {
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

			if code == ptcp.Code_CloseChannel {
				id := msg.Close().Id()
				c.channels.remove(id)
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

func (c *connSocket) close() status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.st.OK() {
		return status.OK
	}

	c.st = statusConnClosed
	c.conn.Close()

	if debug {
		debugPrint(c.client, "conn.close\t", c.st)
	}
	return c.st
}

func (c *connSocket) closed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return !c.st.OK()
}

func (c *connSocket) status() status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.st
}

// channels

type connChannels struct {
	mu       sync.Mutex
	client   bool
	closed   bool
	channels map[bin.Bin128]*channel
}

func newConnChannels(client bool) connChannels {
	return connChannels{
		client:   client,
		channels: make(map[bin.Bin128]*channel),
	}
}

func (c *connChannels) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}
	c.closed = true

	for _, ch := range c.channels {
		ch.connClosed()
	}
}

func (c *connChannels) get(id bin.Bin128) (*channel, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, false
	}

	st, ok := c.channels[id]
	return st, ok
}

func (c *connChannels) open(conn *conn) (*channel, status.Status) {
	id := bin.Random128()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, statusConnClosed
	}

	if debug {
		debugPrint(c.client, "conn.open\t", id)
	}

	ch := openChannel(id, conn)
	c.channels[ch.id] = ch
	return ch, status.OK
}

func (c *connChannels) opened(conn *conn, id bin.Bin128) status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return statusConnClosed
	}

	_, ok := c.channels[id]
	if ok {
		return tcpErrorf("ch %v already exists", id) // impossible
	}

	if debug {
		debugPrint(c.client, "conn.opened\t", id)
	}

	ch := openedChannel(id, conn)
	c.channels[id] = ch
	return status.OK
}

func (c *connChannels) remove(id bin.Bin128) (*channel, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, false
	}

	st, ok := c.channels[id]
	if !ok {
		return nil, false
	}

	delete(c.channels, id)

	if debug {
		debugPrint(c.client, "conn.remove\t", id)
	}
	return st, true
}
