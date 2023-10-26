package tcp

import (
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

	// Methods

	// Connect tries to connect to the server.
	Connect(cancel <-chan struct{}) async.Future[struct{}]

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

var _ Client = (*client)(nil)

type client struct {
	address string
	logger  logging.Logger
	options Options

	lock   async.Lock
	closed bool

	// conn and connecting are mutually exclusive
	conn       *conn
	connected  *async.Flag
	connecting async.Future[struct{}]
}

func newClient(address string, logger logging.Logger, opts Options) *client {
	return &client{
		address: address,
		logger:  logger,
		options: opts.clean(),

		lock:      async.NewLock(),
		connected: async.UnsetFlag(),
	}
}

// Options returns the client options.
func (c *client) Options() Options {
	return c.options
}

// Connected indicates that the client is connected to the server.
func (c *client) Connected() <-chan struct{} {
	return c.connected.Wait()
}

// Methods

// Connect tries to connect to the server.
func (c *client) Connect(cancel <-chan struct{}) async.Future[struct{}] {
	conn, future, st := c.connect(cancel)
	switch {
	case !st.OK():
		return async.Rejected[struct{}](st)
	case conn != nil:
		return async.Resolved[struct{}](struct{}{})
	}
	return future
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
		c.connected.Unset()

		c.conn = nil
	}

	// Cancel connecting
	if c.connecting != nil {
		c.connecting.Cancel()
	}
	return status.OK
}

// Channel returns a new channel.
func (c *client) Channel(cancel <-chan struct{}) (Channel, status.Status) {
	wait := true

	for {
		// Get connection or try to connect
		conn, future, st := c.connect(cancel)
		switch {
		case !st.OK():
			return nil, st
		case conn != nil:
			return conn.Channel(cancel)
		}

		// Return if already tried
		if !wait {
			return nil, status.Unavailable("cannot establish connection, server unavailable")
		}
		wait = false

		// Await connecting
		select {
		case <-cancel:
			return nil, status.Cancelled
		case <-future.Wait():
			if st := future.Status(); !st.OK() {
				return nil, st
			}
		}
	}
}

// private

func (c *client) connect(cancel <-chan struct{}) (*conn, async.Future[struct{}], status.Status) {
	select {
	case <-c.lock:
	case <-cancel:
		return nil, nil, status.Cancelled
	}
	defer c.lock.Unlock()

	// Check closed
	if c.closed {
		return nil, nil, statusClientClosed
	}

	// Check connected
	if c.conn != nil {
		closed := c.conn.closed()
		if !closed {
			return c.conn, nil, status.OK
		}

		c.conn.Free()
		c.connected.Unset()
		c.conn = nil
	}

	// Check connecting
	if c.connecting != nil {
		return nil, c.connecting, status.OK
	}

	// Connect
	c.connecting = async.Go(c.dial)
	return nil, c.connecting, status.OK
}

func (c *client) dial(cancel <-chan struct{}) status.Status {
	conn, st := c.dialRecover()

	// Close on error
	ok := false
	defer func() {
		if !ok {
			if conn != nil {
				conn.close()
			}
		}
	}()

	// Handle result
	c.lock.Lock()
	defer c.lock.Unlock()
	c.connecting = nil

	switch {
	case !st.OK():
		c.logger.ErrorStatus("Connection failed", st, "address", c.address)
		return status.OK
	case c.closed:
		return statusClientClosed
	}

	c.conn = conn
	c.connected.Set()
	c.logger.Debug("Connected", "address", c.address)

	ok = true
	return status.OK
}

func (c *client) dialRecover() (_ *conn, st status.Status) {
	defer func() {
		if e := recover(); e != nil {
			st = status.Recover(e)
		}
	}()

	return connect(c.address, c.logger, c.options)
}
