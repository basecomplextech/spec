package tcp

import (
	"sync"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
)

// Client is a SpecTCP client which manages outgoing connections.
type Client interface {
	// Options returns the client options.
	Options() Options

	// Connected indicates that the client is connected to the server.
	Connected() <-chan struct{}

	// Disconnected indicates that the client is disconnected from the server.
	Disconnected() <-chan struct{}

	// IsConnected returns true if the client is connected to the server.
	IsConnected() bool

	// Methods

	// Connect manually starts the internal connect loop.
	Connect() status.Status

	// Close closes the client.
	Close() status.Status

	// Channel returns a new channel.
	Channel(cancel <-chan struct{}) (Channel, status.Status)
}

// NewClient returns a new client.
func NewClient(address string, logger logging.Logger, opts Options) Client {
	return newClient(address, logger, opts)
}

// internal

const (
	minConnectRetryTimeout = time.Millisecond * 25
	maxConnectRetryTimeout = time.Second
)

var _ Client = (*client)(nil)

type client struct {
	address string
	logger  logging.Logger
	options Options

	closed_       *async.Flag
	connected_    *async.Flag
	disconnected_ *async.Flag

	mu        sync.Mutex
	closed    bool
	conn      *conn
	connector async.Routine[struct{}]
}

func newClient(address string, logger logging.Logger, opts Options) *client {
	return &client{
		address: address,
		logger:  logger,
		options: opts.clean(),

		closed_:       async.UnsetFlag(),
		connected_:    async.UnsetFlag(),
		disconnected_: async.SetFlag(),
	}
}

// Options returns the client options.
func (c *client) Options() Options {
	return c.options
}

// Connected indicates that the client is connected to the server.
func (c *client) Connected() <-chan struct{} {
	return c.connected_.Wait()
}

// Disconnected indicates that the client is disconnected from the server.
func (c *client) Disconnected() <-chan struct{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.disconnected()
	}
	return c.disconnected_.Wait()
}

// IsConnected returns true if the client is connected to the server.
func (c *client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return false
	}

	closed := c.conn.closed()
	return !closed
}

// Methods

// Connect manually starts the internal connect loop.
func (c *client) Connect() status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch {
	case c.closed:
		return statusClientClosed
	case c.connector != nil:
		return status.OK
	}

	c.connector = async.Go(c.connect)
	return status.OK
}

// Close closes the client.
func (c *client) Close() status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check closed
	if c.closed {
		return status.OK
	}

	c.closed = true
	c.closed_.Set()
	c.connected_.Unset()
	c.disconnected_.Set()

	// Close connection
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	// Stop connector
	if c.connector != nil {
		c.connector.Cancel()
		c.connector = nil
	}
	return status.OK
}

// Channel returns a new channel.
func (c *client) Channel(cancel <-chan struct{}) (Channel, status.Status) {
	conn, st := c.awaitConn(cancel)
	if !st.OK() {
		return nil, st
	}
	return conn.Channel(cancel)
}

// private

// awaitConn returns a connection or waits for it, starts the connector if not running.
func (c *client) awaitConn(cancel <-chan struct{}) (*conn, status.Status) {
	for {
		conn, st := c.getConn()
		switch st.Code {
		default:
			return nil, st
		case status.CodeOK:
			return conn, status.OK
		case status.CodeUnavailable:
		}

		select {
		case <-cancel:
			return nil, status.Cancelled
		case <-c.closed_.Wait():
			return nil, statusClientClosed
		case <-c.connected_.Wait():
		}
	}
}

// getConn returns a connection if present, starts the connector if not running.
func (c *client) getConn() (*conn, status.Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check closed
	if c.closed {
		return nil, statusClientClosed
	}

	// Check connection
	if c.conn != nil {
		if !c.conn.closed() {
			return c.conn, status.OK
		}
		c.conn = nil
	}

	// Check connector running
	if c.connector == nil {
		c.connector = async.Go(c.connect)
	}

	return nil, status.Unavailable("client not connected")
}

// connect runs the connect loop.
func (c *client) connect(cancel <-chan struct{}) status.Status {
	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		c.connector = nil
	}()

	for {
		st := c.doConnect(cancel)
		switch st.Code {
		case status.CodeOK:
		case status.CodeCancelled:
			return st
		case status.CodeUnavailable:
		default:
			c.logger.ErrorStatus("Internal client error", st)
		}
	}
}

func (c *client) doConnect(cancel <-chan struct{}) (st status.Status) {
	defer func() {
		if e := recover(); e != nil {
			st = status.Recover(e)
		}
	}()

	conn := (*conn)(nil)
	timer := (*time.Timer)(nil)
	timeout := minConnectRetryTimeout
	logged := false

	// Connect/retry
	for {
		select {
		default:
		case <-cancel:
			return status.Cancelled
		}

		// Connect
		if c.logger.TraceEnabled() {
			c.logger.Trace("Client connecting...", "address", c.address)
		}
		conn, st = connect(c.address, c.logger, c.options)
		if st.OK() {
			break
		}

		// Exponential backoff
		if timer != nil {
			timeout = timeout * 2
			if timeout > maxConnectRetryTimeout {
				timeout = maxConnectRetryTimeout
			}
		}

		// Log error once
		switch {
		case !logged:
			logged = true
			c.logger.ErrorStatus("Client connection failed", st, "address", c.address)
		case c.logger.TraceEnabled():
			c.logger.Trace("Client connection failed", "status", st, "address", c.address,
				"retry", timeout)
		}

		// Start or reset timer
		if timer == nil {
			timer = time.NewTimer(timeout)
		} else {
			timer.Reset(timeout)
		}

		// Await timeout, cancel or close
		select {
		case <-cancel:
			return status.Cancelled
		case <-c.closed_.Wait():
			return status.Cancelled
		case <-timer.C:
			// Continue
		}
	}

	// Disconnect on exit
	defer conn.Free()
	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		c.conn = nil
		c.connected_.Unset()
		c.disconnected_.Set()
		c.logger.Debug("Client disconnected", "address", c.address)
	}()

	// Connected
	{
		c.mu.Lock()
		if c.closed {
			c.mu.Unlock()
			return status.Cancelled
		}

		c.conn = conn
		c.connected_.Set()
		c.disconnected_.Unset()
		c.mu.Unlock()

		c.logger.Debug("Client connected", "address", c.address)
	}

	// Await cancel/close/disconnect
	select {
	case <-cancel:
		return status.Cancelled
	case <-c.closed_.Wait():
		return status.Cancelled
	case <-conn.disconnected():
		return status.OK
	}
}
