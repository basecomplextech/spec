package tcp

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/ptcp"
)

type Conn interface {
	// Open opens a new stream and immediately sends a message.
	Open(cancel <-chan struct{}, msg []byte) (Stream, status.Status)
}

// internal

var _ Conn = (*conn)(nil)

type conn struct {
	conn    net.Conn
	server  bool
	handler Handler // only for server connections

	r     *bufio.Reader
	rhead [4]byte
	rbody *alloc.Buffer

	w     *bufio.Writer
	whead [4]byte

	mu      sync.RWMutex
	st      status.Status
	streams map[bin.Bin128]*stream

	writeMu    sync.Mutex
	writeBuf   *alloc.Buffer
	writeMsg   spec.Writer
	writeQueue *queue
}

func newConn(c net.Conn) *conn {
	buf := alloc.NewBuffer()

	return &conn{
		conn: c,

		r:     bufio.NewReader(c),
		rbody: alloc.NewBuffer(),

		w: bufio.NewWriter(c),

		st:      status.OK,
		streams: make(map[bin.Bin128]*stream),

		writeBuf:   buf,
		writeMsg:   spec.NewWriterBuffer(buf),
		writeQueue: newQueue(),
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

func (c *conn) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.st.OK() {
		return
	}
	c.st = status.New(codeClosed, "connection closed")
	c.conn.Close()

	for _, s := range c.streams {
		s.close()
	}
	c.streams = nil
}

func (c *conn) run(cancel <-chan struct{}) status.Status {
	defer c.close() // for panics

	reading := async.Go(c.readLoop)
	writing := async.Go(c.writeLoop)
	defer async.CancelWaitAll(reading, writing)

	var st status.Status
	select {
	case <-cancel:
		st = status.Cancelled
	case <-reading.Wait():
		st = reading.Status()
		log.Fatal(st)
	case <-writing.Wait():
		st = writing.Status()
		log.Fatal(st)
	}

	c.close()
	return st
}

// write loop

func (c *conn) writeLoop(cancel <-chan struct{}) status.Status {
	defer c.writeBuf.Free()
	defer c.writeQueue.free()
	head := c.whead[:]

	for {
		// Write message
		msg, ok := c.writeQueue.next()
		if ok {
			size := len(msg)
			binary.BigEndian.PutUint32(head, uint32(size))

			if _, err := c.w.Write(head); err != nil {
				return tcpError(err)
			}
			if _, err := c.w.Write(msg); err != nil {
				return tcpError(err)
			}

			continue
		}

		// Flush if no more messages
		if err := c.w.Flush(); err != nil {
			return status.WrapError(err)
		}

		// Wait for more messages
		select {
		case <-cancel:
			return status.Cancelled
		case <-c.writeQueue.wait():
			continue
		}
	}
}

// read loop

func (c *conn) readLoop(cancel <-chan struct{}) status.Status {
	head := c.rhead[:]

	for {
		// Read size
		if _, err := io.ReadFull(c.r, head); err != nil {
			return tcpError(err)
		}
		size := binary.BigEndian.Uint32(head)

		// Read body
		c.rbody.Reset()
		body := c.rbody.Grow(int(size))
		if _, err := io.ReadFull(c.r, body); err != nil {
			return tcpError(err)
		}

		// Parse message
		msg, _, err := ptcp.ParseMessage(body)
		if err != nil {
			return tcpError(err)
		}

		// Handle message
		if st := c.handleMessage(msg); !st.OK() {
			return st
		}
	}
}

// handleMessage handles a message.
func (c *conn) handleMessage(msg ptcp.Message) status.Status {
	code := msg.Code()
	switch code {
	case ptcp.Code_OpenStream:
		return c.handleOpenStream(msg.Open())
	case ptcp.Code_CloseStream:
		return c.handleCloseStream(msg.Close())
	case ptcp.Code_StreamMessage:
		return c.handleStreamMessage(msg.Message())
	}
	return status.OK
}

// handleOpenStream handles an open stream message.
func (c *conn) handleOpenStream(msg ptcp.OpenStream) status.Status {
	// Create stream
	s, ok, st := c._handleOpenStream(msg)
	switch {
	case !st.OK():
		return st
	case !ok:
		return status.OK
	}

	// Start stream
	go c.handler.Handle(s)

	// Receive message
	data := msg.Data()
	if len(data) > 0 {
		s.receive(data)
	}
	return status.OK
}

func (c *conn) _handleOpenStream(msg ptcp.OpenStream) (*stream, bool, status.Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check conn open
	if !c.st.OK() {
		return nil, false, status.OK
	}

	// Check stream exists
	id := msg.Id()
	_, ok := c.streams[id]
	if ok {
		// TODO: Send error
		return nil, false, status.OK
	}

	// Create stream
	s := newStream(c, id)
	c.streams[id] = s
	return s, true, status.OK
}

// handleCloseStream handles a close stream message.
func (c *conn) handleCloseStream(msg ptcp.CloseStream) status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check conn open
	if !c.st.OK() {
		return status.OK
	}

	// Ignore absent stream
	id := msg.Id()
	s, ok := c.streams[id]
	if !ok {
		return status.OK
	}

	// Delete stream
	delete(c.streams, id)

	// Close stream
	s.close()
	return status.OK
}

// handleStreamMessage handles a stream message.
func (c *conn) handleStreamMessage(msg ptcp.StreamMessage) status.Status {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check conn open
	if !c.st.OK() {
		return status.OK
	}

	// Check stream exists
	id := msg.Id()
	s, ok := c.streams[id]
	if !ok {
		// TODO: Send error
		return status.OK
	}

	// Receive data
	data := msg.Data()
	s.receive(data)
	return status.OK
}

// streams

// openStream opens a new stream and immediately sends a message.
func (c *conn) openStream(b []byte) (Stream, status.Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check conn open
	if !c.st.OK() {
		return nil, c.st
	}

	// Create stream
	id := bin.Random128()
	s := newStream(c, id)
	c.streams[id] = s

	// Write message
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// TODO: Close stream on error
	var msg ptcp.Message
	{
		c.writeBuf.Reset()
		c.writeMsg.Reset(c.writeBuf)

		w := ptcp.NewMessageWriterTo(c.writeMsg.Message())
		w.Code(ptcp.Code_OpenStream)

		w1 := w.Open()
		w1.Id(id)
		w1.Data(b)
		if err := w1.End(); err != nil {
			return nil, tcpError(err)
		}

		var err error
		msg, err = w.Build()
		if err != nil {
			return nil, tcpError(err)
		}
	}

	data := msg.Unwrap().Raw()
	c.writeQueue.append(data)
	return s, status.OK
}

// streamClose is called by a stream to signal a close.
func (c *conn) streamClose(id bin.Bin128) status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check conn open
	if !c.st.OK() {
		return c.st
	}

	// Check stream open
	_, ok := c.streams[id]
	if !ok {
		return status.OK
	}

	// Delete stream
	delete(c.streams, id)

	// Write message
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	var msg ptcp.Message
	{
		c.writeBuf.Reset()
		c.writeMsg.Reset(c.writeBuf)

		w := ptcp.NewMessageWriterTo(c.writeMsg.Message())
		w.Code(ptcp.Code_CloseStream)

		w1 := w.Close()
		w1.Id(id)
		if err := w1.End(); err != nil {
			return tcpError(err)
		}

		var err error
		msg, err = w.Build()
		if err != nil {
			return tcpError(err)
		}
	}

	b := msg.Unwrap().Raw()
	c.writeQueue.append(b)
	return status.OK
}

// streamSend is called by a stream to send a message.
func (c *conn) streamSend(id bin.Bin128, b []byte) (bool, status.Status) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check conn open
	if !c.st.OK() {
		return false, c.st
	}

	// Check stream closed
	_, ok := c.streams[id]
	if !ok {
		return false, status.OK
	}

	// Write message
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	var msg ptcp.Message
	{
		c.writeBuf.Reset()
		c.writeMsg.Reset(c.writeBuf)

		w := ptcp.NewMessageWriterTo(c.writeMsg.Message())
		w.Code(ptcp.Code_StreamMessage)

		w1 := w.Message()
		w1.Id(id)
		w1.Data(b)
		if err := w1.End(); err != nil {
			return false, tcpError(err)
		}

		var err error
		msg, err = w.Build()
		if err != nil {
			return false, tcpError(err)
		}
	}

	data := msg.Unwrap().Raw()
	c.writeQueue.append(data)
	return true, status.OK
}
