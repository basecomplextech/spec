// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"math/rand/v2"
	"sync"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/collect/slices2"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/opt"
	"github.com/basecomplextech/baselibrary/status"
)

// Client is a SpecMPX client which manages outgoing connections.
type Client interface {
	// Address returns the server address.
	Address() string

	// Options returns the client options.
	Options() Options

	// Flags

	// Closed indicates that the client is closed.
	Closed() async.Flag

	// Connected indicates that the client is connected to the server.
	Connected() async.Flag

	// Disconnected indicates that the client is disconnected from the server.
	Disconnected() async.Flag

	// Lifecycle

	// Close closes the client.
	Close() status.Status

	// Methods

	// Conn returns a connection.
	Conn(ctx async.Context) (Conn, status.Status)

	// Channel returns a new channel.
	Channel(ctx async.Context) (Channel, status.Status)
}

// NewClient returns a new client.
func NewClient(addr string, logger logging.Logger, opts Options) Client {
	return newClient(addr, logger, opts)
}

// internal

var _ Client = (*client)(nil)

type client struct {
	addr    string
	logger  logging.Logger
	options Options

	closed_       async.MutFlag
	connected_    async.MutFlag
	disconnected_ async.MutFlag

	mu         sync.Mutex
	conns      []conn
	connecting opt.Opt[async.Routine[conn]]
}

func newClient(addr string, logger logging.Logger, opts Options) *client {
	c := &client{
		addr:    addr,
		logger:  logger,
		options: opts.clean(),

		closed_:       async.UnsetFlag(),
		connected_:    async.UnsetFlag(),
		disconnected_: async.SetFlag(),
	}

	if opts.Client.AutoConnect {
		c.connect()
	}
	return c
}

// Address returns the server address.
func (c *client) Address() string {
	return c.addr
}

// Options returns the client options.
func (c *client) Options() Options {
	return c.options
}

// Flags

// Closed indicates that the client is closed.
func (c *client) Closed() async.Flag {
	return c.closed_
}

// Connected indicates that the client is connected to the server.
func (c *client) Connected() async.Flag {
	return c.connected_
}

// Disconnected indicates that the client is disconnected from the server.
func (c *client) Disconnected() async.Flag {
	return c.disconnected_
}

// Lifecycle

// Close closes the client.
func (c *client) Close() status.Status {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed_.Get() {
		return status.OK
	}
	c.closed_.Set()

	// Stop connecting
	if routine, ok := c.connecting.Unwrap(); ok {
		c.connecting.Unset()
		routine.Stop()
	}

	// Close connections
	for _, conn := range c.conns {
		conn.Close()
	}
	c.conns = nil

	// Update flags
	c.connected_.Unset()
	c.disconnected_.Set()
	return status.OK
}

// Methods

// Conn returns a connection.
func (c *client) Conn(ctx async.Context) (Conn, status.Status) {
	// Get connection
	conn, future, st := c.conn()
	if !st.OK() {
		return nil, st
	}
	if conn != nil {
		return conn, status.OK
	}

	// Await connection
	select {
	case <-ctx.Wait():
		return nil, ctx.Status()
	case <-future.Wait():
		return future.Result()
	}
}

// Channel returns a new channel.
func (c *client) Channel(ctx async.Context) (Channel, status.Status) {
	conn, st := c.Conn(ctx)
	if !st.OK() {
		return nil, st
	}
	return conn.Channel(ctx)
}

// conn delegate

// onConnClosed is called when the connection is closed.
func (c *client) onConnClosed(conn conn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Delete connection
	c.conns = slices2.Remove(c.conns, conn)
	if len(c.conns) > 0 {
		return
	}

	// Clear connected
	if c.connected_.Get() {
		c.connected_.Unset()
		c.disconnected_.Set()
	}

	// Maybe auto-connect
	if c.options.Client.AutoConnect {
		c.connect()
	}
}

// onConnChannelsReached is called when the number of channels reaches the target.
// The method is used by the auto connector to establish more connections.
func (c *client) onConnChannelsReached(conn conn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	max := c.options.Client.MaxConns
	if max <= 0 {
		return
	}

	num := len(c.conns)
	if num < max {
		c.connect()
	}
}

// private

func (c *client) conn() (conn, async.Future[conn], status.Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check closed
	if c.closed_.Get() {
		return nil, nil, status.Closedf("mpx client closed")
	}

	// Round-robin connections
	for len(c.conns) > 0 {
		i := rand.IntN(len(c.conns))
		conn := c.conns[i]
		closed := conn.Closed().Get()
		if closed {
			c.conns = slices2.RemoveAt(c.conns, i, 1)
			continue
		}
		return conn, nil, status.OK
	}

	// Maybe clear connected
	if c.connected_.Get() {
		c.connected_.Unset()
		c.disconnected_.Set()
	}

	// Return new connection
	future, st := c.connect()
	if !st.OK() {
		return nil, nil, st
	}
	return nil, future, status.OK
}

func (c *client) connect() (async.Future[conn], status.Status) {
	routine, ok := c.connecting.Unwrap()
	if ok {
		return routine, status.OK
	}

	routine = async.Run(c.doConnect)
	c.connecting.Set(routine)
	return routine, status.OK
}

func (c *client) doConnect(ctx async.Context) (conn, status.Status) {
	// Clear routine on exit
	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		c.connecting.Unset()
	}()

	// Connect
	conn, st := connect(c.addr, c /* delegate */, c.logger, c.options)
	if !st.OK() {
		return nil, st
	}

	// Add connection
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed_.Get() {
		if conn != nil {
			conn.Close()
		}
		return nil, status.Closedf("mpx client closed")
	}

	c.conns = append(c.conns, conn)
	c.connected_.Set()
	c.disconnected_.Unset()
	return conn, status.OK
}
