package tcp

import (
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
)

// Client is a SpecTCP client which manages outgoing connections.
type Client interface {
	// Close closes the client.
	Close() status.Status

	// Connect connects to the server and returns a connection.
	Connect(cancel <-chan struct{}) (Conn, status.Status)
}

// NewClient returns a new client.
func NewClient(address string, logger logging.Logger) Client {
	return newClient(address, logger)
}

// internal

var _ Client = (*client)(nil)

type client struct {
	address string
	logger  logging.Logger

	lock   async.Lock
	conn   *ref.R[*conn]
	closed bool
}

func newClient(address string, logger logging.Logger) *client {
	return &client{
		address: address,
		logger:  logger,

		lock: async.NewLock(),
	}
}

// Close closes the client.
func (c *client) Close() status.Status {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Check closed
	if c.closed {
		return status.OK
	}
	c.closed = true

	// Close connection
	if c.conn != nil {
		conn := c.conn.Unwrap()
		conn.close()

		c.conn.Release()
		c.conn = nil
	}
	return status.OK
}

// Connect connects to the server and returns a connection.
func (c *client) Connect(cancel <-chan struct{}) (Conn, status.Status) {
	for {
		conn, st := c.connect(cancel)
		if !st.OK() {
			return nil, st
		}

		closed := conn.Unwrap().closed()
		if closed {
			conn.Unwrap().Free()
			continue
		}

		return newClientConn(conn), status.OK
	}
}

// private

func (c *client) connect(cancel <-chan struct{}) (*ref.R[*conn], status.Status) {
	select {
	case <-c.lock:
	case <-cancel:
		return nil, status.Cancelled
	}
	defer c.lock.Unlock()

	// Check closed
	if c.closed {
		return nil, statusClientClosed
	}

	// Check existing connection
	if c.conn != nil {
		closed := c.conn.Unwrap().closed()
		if !closed {
			return ref.Retain(c.conn), status.OK
		}

		c.conn.Release()
		c.conn = nil
	}

	// Open new connection
	conn, st := connect(c.address, c.logger)
	if !st.OK() {
		return nil, st
	}

	c.conn = ref.New(conn)
	c.logger.Debug("Connected", "address", c.address)
	return ref.Retain(c.conn), status.OK
}

// client conn

var _ Conn = (*clientConn)(nil)

type clientConn struct {
	conn  *ref.R[*conn]
	freed bool
}

func newClientConn(conn *ref.R[*conn]) *clientConn {
	return &clientConn{conn: conn}
}

// Close closes the connection.
func (c *clientConn) Close() status.Status {
	// Ignore, the connection will be closed on release.
	return status.OK
}

// Stream opens a new stream.
func (c *clientConn) Stream(cancel <-chan struct{}) (Stream, status.Status) {
	return c.conn.Unwrap().Stream(cancel)
}

// Free closes and frees the connection.
func (c *clientConn) Free() {
	if c.freed {
		return
	}

	c.freed = true
	c.conn.Release()
	c.conn = nil
}
