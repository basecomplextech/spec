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

	reader *reader
	writer *writer
	writeq alloc.MQueue

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

		reader: newReader(c, client),
		writer: newWriter(c, client),
		writeq: alloc.NewMQueueCap(connQueueCap),
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
	defer c.writeq.Free()

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

	// Connect
	if st := c.connect(cancel); !st.OK() {
		c.logger.Error("Connection failed", "client", c.client, "status", st)
		return st
	}

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
	defer c.writeq.Close()

	c.socket.close()
}

func (c *conn) closed() bool {
	return c.socket.closed()
}

// connect

func (c *conn) connect(cancel <-chan struct{}) status.Status {
	if c.client {
		return c.connectClient(cancel)
	} else {
		return c.connectServer(cancel)
	}
}

func (c *conn) connectClient(cancel <-chan struct{}) status.Status {
	// Write protocol line
	if st := c.writer.writeString(ProtocolLine); !st.OK() {
		return st
	}

	// Write connect request
	{
		w := ptcp.NewConnectRequestWriter()
		vv := w.Versions()
		vv.Add(ptcp.Version_Version10)
		vv.End()

		req, err := w.Build()
		if err != nil {
			return tcpError(err)
		}

		if st := c.writer.writeRequest(req); !st.OK() {
			return st
		}
	}

	// Read/check protocol line
	line, st := c.reader.readLine()
	if !st.OK() {
		return st
	}
	if line != ProtocolLine {
		return tcpErrorf("invalid protocol, expected %q, got %q", ProtocolLine, line)
	}

	// Read connect response
	resp, st := c.reader.readResponse()
	if !st.OK() {
		return st
	}

	// Check status
	ok := resp.Ok()
	if !ok {
		return tcpErrorf("server refused connection: %v", resp.Error())
	}

	// Check version
	v := resp.Version()
	if v != ptcp.Version_Version10 {
		return tcpErrorf("server returned unsupported version %d", v)
	}
	return status.OK
}

func (c *conn) connectServer(cancel <-chan struct{}) status.Status {
	// Write protocol line
	if st := c.writer.writeString(ProtocolLine); !st.OK() {
		return st
	}

	// Read/check protocol line
	line, st := c.reader.readLine()
	if !st.OK() {
		return st
	}
	if line != ProtocolLine {
		return tcpErrorf("invalid protocol, expected %q, got %q", ProtocolLine, line)
	}

	// Read connect request
	req, st := c.reader.readRequest()
	if !st.OK() {
		return st
	}

	// Check version
	ok := false
	vv := req.Versions()
	for i := 0; i < vv.Len(); i++ {
		v := vv.Get(i)
		if v == ptcp.Version_Version10 {
			ok = true
			break
		}
	}
	if !ok {
		w := ptcp.NewConnectResponseWriter()
		w.Ok(false)
		w.Error("unsupported protocol versions")

		resp, err := w.Build()
		if err != nil {
			return tcpError(err)
		}
		return c.writer.writeResponse(resp)
	}

	// Return response
	w := ptcp.NewConnectResponseWriter()
	w.Ok(true)
	w.Version(ptcp.Version_Version10)

	resp, err := w.Build()
	if err != nil {
		return tcpError(err)
	}
	return c.writer.writeResponse(resp)
}

// read

func (c *conn) readLoop(cancel <-chan struct{}) status.Status {
	for {
		// Receive message
		msg, st := c.reader.readMessage()
		if !st.OK() {
			return st
		}

		// Handle message
		code := msg.Code()
		switch code {
		case ptcp.Code_OpenChannel:
			m := msg.Open()

			if st := c.channels.opened(c, m); !st.OK() {
				return st
			}

		case ptcp.Code_CloseChannel:
			m := msg.Close()
			id := m.Id()

			ch, ok := c.channels.remove(id)
			if !ok {
				continue
			}
			ch.receiveMessage(cancel, msg)

		case ptcp.Code_ChannelMessage:
			m := msg.Message()
			id := m.Id()

			ch, ok := c.channels.get(id)
			if !ok {
				continue
			}
			ch.receiveMessage(cancel, msg)

		case ptcp.Code_ChannelWindow:
			m := msg.Window()
			id := m.Id()

			ch, ok := c.channels.get(id)
			if !ok {
				continue
			}
			ch.receiveWindow(cancel, msg)

		default:
			return tcpErrorf("unexpected tpc message code %d", code)
		}
	}
}

// write

func (c *conn) writeLoop(cancel <-chan struct{}) status.Status {
	for {
		b, ok, st := c.writeq.Read()
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

			if st := c.writer.writeMessage(b); !st.OK() {
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
		case <-c.writeq.ReadWait():
		}
	}
}

// write pushes an outgoing message to the write queue, or returns a connection closed error.
func (c *conn) write(cancel <-chan struct{}, msg ptcp.Message) status.Status {
	b := msg.Unwrap().Raw()

	for {
		ok, st := c.writeq.Write(b)
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
		case <-c.writeq.WriteWait(len(b)):
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

	ch := openChannel(conn, id)
	c.channels[ch.id] = ch
	return ch, status.OK
}

func (c *connChannels) opened(conn *conn, msg ptcp.OpenChannel) status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return statusConnClosed
	}

	id := msg.Id()
	_, ok := c.channels[id]
	if ok {
		return tcpErrorf("ch %v already exists", id) // impossible
	}

	if debug {
		debugPrint(c.client, "conn.opened\t", id)
	}

	ch := openedChannel(conn, msg)
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
