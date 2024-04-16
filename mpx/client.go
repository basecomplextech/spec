package mpx

import (
	"sync"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
)

// Client is a SpecMPX client which manages outgoing connections.
type Client interface {
	// Address returns the server address.
	Address() string

	// Options returns the client options.
	Options() Options

	// IsConnected returns true if the client is connected to the server.
	IsConnected() bool

	// Async

	// Changed adds a connected/disconnected listener.
	// TODO: Maybe remove, obsole.
	Changed() (<-chan struct{}, func())

	// Connected indicates that the client is connected to the server.
	Connected() async.Flag

	// Disconnected indicates that the client is disconnected from the server.
	Disconnected() async.Flag

	// Methods

	// Connect manually starts the internal connect loop.
	Connect() status.Status

	// Close closes the client.
	Close() status.Status

	// Channel returns a new channel.
	Channel(ctx async.Context) (Channel, status.Status)
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

	closed_       async.MutFlag
	connected_    async.MutFlag
	disconnected_ async.MutFlag

	mu     sync.Mutex
	closed bool
	conn   *conn

	connector async.Routine[struct{}]
	listeners map[chan struct{}]struct{}
}

func newClient(address string, logger logging.Logger, opts Options) *client {
	return &client{
		address: address,
		logger:  logger,
		options: opts.clean(),

		closed_:       async.UnsetFlag(),
		connected_:    async.UnsetFlag(),
		disconnected_: async.SetFlag(),

		listeners: make(map[chan struct{}]struct{}),
	}
}

// Address returns the server address.
func (c *client) Address() string {
	return c.address
}

// Options returns the client options.
func (c *client) Options() Options {
	return c.options
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

// Async

// Changed adds a connected/disconnected listener.
func (c *client) Changed() (<-chan struct{}, func()) {
	c.mu.Lock()
	defer c.mu.Unlock()

	l := make(chan struct{}, 1)
	c.listeners[l] = struct{}{}

	unsub := func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.listeners, l)
	}
	return l, unsub
}

// Connected indicates that the client is connected to the server.
func (c *client) Connected() async.Flag {
	return c.connected_
}

// Disconnected indicates that the client is disconnected from the server.
func (c *client) Disconnected() async.Flag {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.disconnected()
	}
	return c.disconnected_
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
func (c *client) Channel(ctx async.Context) (Channel, status.Status) {
	conn, st := c.awaitConn(ctx)
	if !st.OK() {
		return nil, st
	}
	return conn.Channel(ctx)
}

// private

// awaitConn returns a connection or waits for it, starts the connector if not running.
func (c *client) awaitConn(ctx async.Context) (*conn, status.Status) {
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
		case <-ctx.Wait():
			return nil, ctx.Status()
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
func (c *client) connect(ctx async.Context) status.Status {
	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		c.connector = nil
	}()

	for {
		st := c.doConnect(ctx)
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

func (c *client) doConnect(ctx async.Context) (st status.Status) {
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
		case <-ctx.Wait():
			return ctx.Status()
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

		// Await timeout, ctx or close
		select {
		case <-ctx.Wait():
			return ctx.Status()
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
		c.logger.Info("Client disconnected", "address", c.address)

		c.connected_.Unset()
		c.disconnected_.Set()
		c.notifyLocked()
	}()

	// Connected
	st = func() status.Status {
		c.mu.Lock()
		defer c.mu.Unlock()

		c.conn = conn
		c.logger.Info("Client connected", "address", c.address)

		c.connected_.Set()
		c.disconnected_.Unset()
		c.notifyLocked()
		return status.OK
	}()
	if !st.OK() {
		return st
	}

	// Await ctx/close/disconnect
	select {
	case <-ctx.Wait():
		return ctx.Status()
	case <-c.closed_.Wait():
		return status.Cancelled
	case <-conn.disconnected().Wait():
		return status.OK
	}
}

func (c *client) notifyLocked() {
	for l := range c.listeners {
		select {
		case l <- struct{}{}:
		default:
		}
	}
}
