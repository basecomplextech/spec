package tcp

import (
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
)

// Client is a SpecTCP client which manages outgoing connections.
type Client interface {
	// Close closes the client.
	Close() status.Status

	// Channel returns a new channel.
	Channel(cancel <-chan struct{}) (Channel, status.Status)

	// Options returns the client options.
	Options() Options
}

// NewClient returns a new client.
func NewClient(address string, logger logging.Logger, opts Options) Client {
	return newClient(address, logger, opts)
}

// internal

var _ Client = (*client)(nil)

type client struct {
	address string
	logger  logging.Logger
	options Options

	lock   async.Lock
	conn   *conn
	closed bool
}

func newClient(address string, logger logging.Logger, opts Options) *client {
	return &client{
		address: address,
		logger:  logger,
		options: opts.clean(),

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
		c.conn.close()
		c.conn = nil
	}
	return status.OK
}

// Channel returns a new channel.
func (c *client) Channel(cancel <-chan struct{}) (Channel, status.Status) {
	select {
	case <-c.lock:
	case <-cancel:
		return nil, status.Cancelled
	}
	defer c.lock.Unlock()

	conn, st := c.connect(cancel)
	if !st.OK() {
		return nil, st
	}
	return conn.Channel(cancel)
}

// Options returns the client options.
func (c *client) Options() Options {
	return c.options
}

// private

func (c *client) connect(cancel <-chan struct{}) (*conn, status.Status) {
	// Check closed
	if c.closed {
		return nil, statusClientClosed
	}

	// Check existing connection
	if c.conn != nil {
		closed := c.conn.closed()
		if !closed {
			return c.conn, status.OK
		}

		c.conn.Free()
		c.conn = nil
	}

	// Open new connection
	conn, st := connect(c.address, c.logger, c.options)
	if !st.OK() {
		return nil, st
	}

	c.conn = conn
	c.logger.Debug("Connected", "address", c.address)
	return conn, status.OK
}
