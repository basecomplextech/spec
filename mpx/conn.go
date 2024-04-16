package mpx

import (
	"net"
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/pmpx"
)

// Conn is a TCP network connection.
type Conn interface {
	// Close closes the connection and frees its internal resources.
	Close() status.Status

	// Channel opens a new channel.
	Channel(ctx async.Context) (Channel, status.Status)

	// Internal

	// Free closes the connection, allows to wrap it into ref.R[Conn].
	Free()
}

// Connect dials an address and returns a connection.
func Connect(address string, logger logging.Logger, opts Options) (Conn, status.Status) {
	return connect(address, logger, opts)
}

// internal

type conn struct {
	handler Handler
	logger  logging.Logger
	options Options

	client   bool
	socket   connSocket
	channels connChannels

	reader *reader
	writer *writer
	writeq alloc.MQueue
}

func connect(address string, logger logging.Logger, opts Options) (*conn, status.Status) {
	opts = opts.clean()

	// Dial address
	nc, err := net.DialTimeout("tcp", address, opts.DialTimeout)
	if err != nil {
		return nil, tcpError(err)
	}

	// Noop incoming handler
	h := HandleFunc(func(ctx async.Context, ch Channel) status.Status {
		return ch.Close()
	})

	// Make connection
	c := newConn(nc, true /* client */, h, logger, opts)
	go func() {
		defer func() {
			if e := recover(); e != nil {
				st := status.Recover(e)
				logger.ErrorStatus("Connection panic", st)
			}
		}()

		c.run()
	}()
	return c, status.OK
}

func newConn(c net.Conn, client bool, handler Handler, logger logging.Logger, opts Options) *conn {
	return &conn{
		handler: handler,
		logger:  logger,
		options: opts.clean(),

		client:   client,
		socket:   newConnSocket(client, c),
		channels: newConnChannels(client),

		reader: newReader(c, client, int(opts.ReadBufferSize)),
		writer: newWriter(c, client, int(opts.WriteBufferSize)),
		writeq: alloc.NewMQueueCap(int(opts.WriteQueueSize)),
	}
}

// Close closes the connection.
func (c *conn) Close() status.Status {
	c.close()
	return status.OK
}

// Channel opens a new channel.
func (c *conn) Channel(ctx async.Context) (Channel, status.Status) {
	return c.channels.open(c)
}

// Internal

// Free closes the connection, allows to wrap it into ref.R[Conn].
func (c *conn) Free() {
	c.close()
}

// internal

func (c *conn) run() {
	defer c.reader.free()
	defer c.writeq.Free()
	defer c.close()

	// Connect
	st := c.connect()
	switch st.Code {
	case status.CodeOK:
	case status.CodeCancelled,
		status.CodeEnd,
		status.CodeClosed:
		return
	default:
		c.logger.ErrorStatus("Connection error", st)
		return
	}

	// Run loops
	reader := async.Go(c.readLoop)
	writer := async.Go(c.writeLoop)
	defer async.CancelWaitAll(reader, writer)
	defer c.close()

	// Wait exit
	select {
	case <-reader.Wait():
		st = reader.Status()
	case <-writer.Wait():
		st = writer.Status()
	}

	// Maybe log error
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeEnd,
		status.CodeClosed:
	default:
		c.logger.ErrorStatus("Connection error", st)
	}
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

func (c *conn) disconnected() async.Flag {
	return c.socket.disconnected
}

// connect

func (c *conn) connect() status.Status {
	if c.client {
		return c.connectClient()
	} else {
		return c.connectServer()
	}
}

func (c *conn) connectClient() status.Status {
	// Write protocol line
	if st := c.writer.writeString(ProtocolLine); !st.OK() {
		return st
	}

	// Write connect request
	{
		w := pmpx.NewConnectRequestWriter()

		vv := w.Versions()
		vv.Add(pmpx.Version_Version10)
		vv.End()

		if c.options.Compress {
			cc := w.Compress()
			cc.Add(pmpx.Compress_Lz4)
			cc.End()
		}

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
	if v != pmpx.Version_Version10 {
		return tcpErrorf("server returned unsupported version %d", v)
	}

	// Init compression
	comp := resp.Compress()
	switch comp {
	case pmpx.Compress_None:
	case pmpx.Compress_Lz4:
		if st := c.reader.initLZ4(); !st.OK() {
			return st
		}
		if st := c.writer.initLZ4(); !st.OK() {
			return st
		}
	default:
		return tcpErrorf("server returned unsupported compression %d", comp)
	}
	return status.OK
}

func (c *conn) connectServer() status.Status {
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
	versions := req.Versions()
	for i := 0; i < versions.Len(); i++ {
		v := versions.Get(i)
		if v == pmpx.Version_Version10 {
			ok = true
			break
		}
	}
	if !ok {
		w := pmpx.NewConnectResponseWriter()
		w.Ok(false)
		w.Error("unsupported protocol versions")

		resp, err := w.Build()
		if err != nil {
			return tcpError(err)
		}
		return c.writer.writeResponse(resp)
	}

	// Select compression
	comp := pmpx.Compress_None
	comps := req.Compress()
	for i := 0; i < comps.Len(); i++ {
		c := comps.Get(i)
		if c == pmpx.Compress_Lz4 {
			comp = pmpx.Compress_Lz4
			break
		}
	}

	// Return response
	{
		w := pmpx.NewConnectResponseWriter()
		w.Ok(true)
		w.Version(pmpx.Version_Version10)
		w.Compress(comp)

		resp, err := w.Build()
		if err != nil {
			return tcpError(err)
		}
		if st := c.writer.writeResponse(resp); !st.OK() {
			return st
		}
	}

	// Init compression
	switch comp {
	case pmpx.Compress_None:
	case pmpx.Compress_Lz4:
		if st := c.reader.initLZ4(); !st.OK() {
			return st
		}
		if st := c.writer.initLZ4(); !st.OK() {
			return st
		}
	}
	return status.OK
}

// read

func (c *conn) readLoop(ctx async.Context) status.Status {
	for {
		// Receive message
		msg, st := c.reader.readMessage()
		if !st.OK() {
			return st
		}

		// Handle message
		code := msg.Code()
		switch code {
		case pmpx.Code_OpenChannel:
			m := msg.Open()

			ch, st := c.channels.opened(c, m)
			if !st.OK() {
				return st
			}
			ch.receiveMessage(ctx, msg)

		case pmpx.Code_CloseChannel:
			m := msg.Close()
			id := m.Id()

			ch, ok := c.channels.remove(id)
			if !ok {
				continue
			}
			ch.receiveMessage(ctx, msg)

		case pmpx.Code_ChannelMessage:
			m := msg.Message()
			id := m.Id()

			ch, ok := c.channels.get(id)
			if !ok {
				continue
			}
			ch.receiveMessage(ctx, msg)

		case pmpx.Code_ChannelWindow:
			m := msg.Window()
			id := m.Id()

			ch, ok := c.channels.get(id)
			if !ok {
				continue
			}
			ch.receiveWindow(ctx, msg)

		default:
			return tcpErrorf("unexpected tpc message code %d", code)
		}
	}
}

// write

func (c *conn) writeLoop(ctx async.Context) status.Status {
	for {
		b, ok, st := c.writeq.Read()
		switch {
		case !st.OK():
			return st

		case ok:
			msg := pmpx.NewMessage(b)
			code := msg.Code()

			if code == pmpx.Code_CloseChannel {
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
		case <-ctx.Wait():
			return ctx.Status()
		case <-c.writeq.ReadWait():
		}
	}
}

// write pushes an outgoing message to the write queue, or returns a connection closed error.
func (c *conn) write(ctx async.Context, msg pmpx.Message) status.Status {
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
		case <-ctx.Wait():
			return ctx.Status()
		case <-c.writeq.WriteWait(len(b)):
			continue
		}
	}
}

// socket

type connSocket struct {
	disconnected async.MutFlag

	mu     sync.Mutex
	st     status.Status
	conn   net.Conn
	client bool
}

func newConnSocket(client bool, conn net.Conn) connSocket {
	return connSocket{
		disconnected: async.UnsetFlag(),

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
	c.disconnected.Set()

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

	ch := openChannel(conn, id, int(conn.options.ChannelWindowSize))
	c.channels[ch.id] = ch
	return ch, status.OK
}

func (c *connChannels) opened(conn *conn, msg pmpx.OpenChannel) (*channel, status.Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, statusConnClosed
	}

	id := msg.Id()
	_, ok := c.channels[id]
	if ok {
		return nil, tcpErrorf("ch %v already exists", id) // impossible
	}

	if debug {
		debugPrint(c.client, "conn.opened\t", id)
	}

	ch := openedChannel(conn, msg)
	c.channels[id] = ch
	return ch, status.OK
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
