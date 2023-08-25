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

	// Open opens a new stream.
	Open(cancel <-chan struct{}) (Stream, status.Status)

	// Request opens a new stream and sends a request message.
	Request(cancel <-chan struct{}, request []byte) (Stream, status.Status)

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

	accept *acceptQueue
	reader *reader
	writer *writer

	mu      sync.RWMutex
	st      status.Status
	main    async.Routine[struct{}]
	streams map[bin.Bin128]*stream

	pendingMu   sync.Mutex
	pending     *pending
	pendingChan chan struct{}
}

func newConn(c net.Conn, client bool, logger logging.Logger) *conn {
	return &conn{
		conn:   c,
		client: client,

		accept: newAcceptQueue(),
		reader: newReader(c, client),
		writer: newWriter(c, client),

		st:      status.OK,
		streams: make(map[bin.Bin128]*stream),

		pending:     newPending(),
		pendingChan: make(chan struct{}, 1),
	}
}

// Accept accepts a new stream, or returns an end status.
func (c *conn) Accept(cancel <-chan struct{}) (Stream, status.Status) {
	for {
		s, ok, st := c.accept.poll()
		switch {
		case !st.OK():
			return nil, st
		case ok:
			return s, status.OK
		}

		select {
		case <-cancel:
			return nil, status.Cancelled
		case <-c.accept.wait():
			continue
		}
	}
}

// Open opens a new stream.
func (c *conn) Open(cancel <-chan struct{}) (Stream, status.Status) {
	return c.open(nil)
}

// Request opens a new stream and sends a request message.
func (c *conn) Request(cancel <-chan struct{}, request []byte) (Stream, status.Status) {
	return c.open(request)
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
	c.accept.close()
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

// streams

func (c *conn) get(id bin.Bin128) (*stream, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	s, ok := c.streams[id]
	return s, ok
}

func (c *conn) open(request []byte) (*stream, status.Status) {
	id := bin.Random128()

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.st.OK() {
		return nil, c.st
	}

	// Check id
	_, ok := c.streams[id]
	if ok {
		return nil, tcpErrorf("failed to generate unique stream id") // impossible
	}

	if debug {
		debugPrint(c.client, "conn.open\t", id)
	}

	// Create stream
	s := openStream(id, c, request)
	c.streams[id] = s
	c.enqueue(id)
	return s, status.OK
}

func (c *conn) remove(id bin.Bin128) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.streams[id]
	if !ok {
		return
	}

	if debug {
		debugPrint(c.client, "conn.remove\t", id)
	}

	delete(c.streams, id)
}

// pending

func (c *conn) enqueue(id bin.Bin128) {
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

func (c *conn) switchPending(next *pending) *pending {
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
			m := msg.Open()
			if st := c.handleOpened(m); !st.OK() {
				return st
			}

		case ptcp.Code_CloseStream:
			m := msg.Close()
			if st := c.handleClosed(m); !st.OK() {
				return st
			}

		case ptcp.Code_StreamMessage:
			m := msg.Message()
			if st := c.handleMessage(m); !st.OK() {
				return st
			}

		default:
			return tcpErrorf("unexpected tpc message code %d", code)
		}
	}
}

func (c *conn) handleOpened(msg ptcp.OpenStream) status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := msg.Id()
	_, ok := c.streams[id]
	if ok {
		return tcpErrorf("stream %v already exists", id) // impossible
	}

	if debug {
		debugPrint(c.client, "conn.opened\t", id)
	}

	data := msg.Data()
	s := openedStream1(id, c, data)
	c.streams[id] = s
	c.accept.push(s)
	c.enqueue(id)
	return status.OK
}

func (c *conn) handleClosed(msg ptcp.CloseStream) status.Status {
	id := msg.Id()

	s, ok := c.get(id)
	if !ok {
		return status.OK
	}

	s.receiveClosed()
	return status.OK
}

func (c *conn) handleMessage(msg ptcp.StreamMessage) status.Status {
	id := msg.Id()

	s, ok := c.get(id)
	if !ok {
		return status.OK
	}

	data := msg.Data()
	s.receiveMessage(data)
	return status.OK
}

// writeLoop

func (c *conn) writeLoop(cancel <-chan struct{}) status.Status {
	active := make(map[bin.Bin128]struct{})
	pending := newPending()

	for {
		for {
			// Add pending streams
			pending = c.switchPending(pending)
			for _, id := range pending.list {
				active[id] = struct{}{}
			}

			// Write active streams
			for id := range active {
				s, ok := c.get(id)
				if !ok {
					delete(active, id)
					continue
				}

				ok, st := s.write(c.writer)
				switch {
				case !st.OK():
					// Write failed
					return st

				case !ok:
					// Stream inactive
					delete(active, id)
				}
			}

			// No more streams, break
			if len(active) == 0 {
				break
			}
		}

		// Flush buffered writes
		if st := c.writer.flush(); !st.OK() {
			return st
		}

		// Wait for active streams
		select {
		case <-cancel:
			return status.Cancelled
		case <-c.pendingChan:
		}
	}
}
