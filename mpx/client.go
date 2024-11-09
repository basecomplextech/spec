// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"math/rand/v2"
	"net"
	"sync"
	"time"

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

	// Conn returns an existing connection, or opens a new one.
	Conn(ctx async.Context) (Conn, status.Status)

	// Channel returns a new channel.
	Channel(ctx async.Context) (Channel, status.Status)
}

// NewClient returns a new client.
func NewClient(addr string, mode ClientMode, logger logging.Logger, opts Options) Client {
	opts = opts.clean()
	return newClient(addr, mode, logger, opts)
}

// NewClientDialer returns a new client with the given dialer.
func NewClientDialer(addr string, mode ClientMode, dialer *net.Dialer, logger logging.Logger,
	opts Options) Client {

	opts = opts.clean()
	return newClientDialer(addr, mode, dialer, logger, opts)
}

// internal

const (
	minConnectRetryTimeout = time.Millisecond * 25
	maxConnectRetryTimeout = time.Second
)

var _ Client = (*client)(nil)

type client struct {
	addr      string
	mode      ClientMode
	connector connector
	logger    logging.Logger
	options   Options

	closed_       async.MutFlag
	connected_    async.MutFlag
	disconnected_ async.MutFlag

	mu    sync.Mutex
	conns []internalConn

	connecting     opt.Opt[async.Routine[internalConn]]
	connectAttempt int // current connect attempt
}

func newClient(addr string, mode ClientMode, logger logging.Logger, opts Options) *client {
	dialer := newDialer(opts)
	return newClientDialer(addr, mode, dialer, logger, opts)
}

func newClientDialer(addr string, mode ClientMode, dialer *net.Dialer, logger logging.Logger,
	opts Options) *client {

	c := &client{
		addr:    addr,
		mode:    mode,
		logger:  logger,
		options: opts,

		closed_:       async.UnsetFlag(),
		connected_:    async.UnsetFlag(),
		disconnected_: async.SetFlag(),
	}
	c.connector = newConnector(dialer, c /* delegate */, logger, opts)

	if mode == ClientMode_AutoConnect {
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

	if c.closed_.IsSet() {
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

// Conn returns an existing connection, or opens a new one.
func (c *client) Conn(ctx async.Context) (Conn, status.Status) {
	// Get connection
	conn, future, st := c.conn()
	if !st.OK() {
		return nil, st
	}
	if conn != nil {
		return conn, status.OK
	}

	// Await connection or dial timeout in auto-connect mode
	// In auto-connect mode, the client will reconnect on errors with exponential backoff.
	// So we we the dial timeout here to avoid waiting too long.
	if c.mode == ClientMode_AutoConnect {
		timeout := c.options.ClientDialTimeout
		if timeout > 0 {
			timer := time.NewTimer(timeout)
			defer timer.Stop()

			select {
			case <-ctx.Wait():
				return nil, ctx.Status()
			case <-future.Wait():
				return future.Result()
			case <-timer.C:
				return nil, status.Timeoutf("mpx dial timeout, address=%v", c.addr)
			}
		}
	}

	// Otherwise, await connection or cancel
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

// connDelegate

var _ connDelegate = (*client)(nil)

// onConnClosed is called when the connection is closed.
func (c *client) onConnClosed(conn internalConn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Delete connection
	c.conns = slices2.Remove(c.conns, conn)
	if len(c.conns) > 0 {
		return
	}

	// Clear connected
	if c.connected_.IsSet() {
		c.connected_.Unset()
		c.disconnected_.Set()
	}

	// Maybe auto-connect
	if c.mode == ClientMode_AutoConnect {
		c.connect()
	}
}

// onConnChannelsReached is called when the number of channels reaches the target.
func (c *client) onConnChannelsReached(conn internalConn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	max := c.options.ClientMaxConns
	if max <= 0 {
		return
	}

	num := len(c.conns)
	if num < max {
		c.connect()
	}
}

// private

func (c *client) conn() (internalConn, async.Future[internalConn], status.Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check closed
	if c.closed_.IsSet() {
		return nil, nil, status.Closedf("mpx client closed")
	}

	// Round-robin connections
	for len(c.conns) > 0 {
		i := rand.IntN(len(c.conns))
		conn := c.conns[i]
		closed := conn.Closed().IsSet()
		if closed {
			c.conns = slices2.RemoveAt(c.conns, i, 1)
			continue
		}
		return conn, nil, status.OK
	}

	// Maybe clear connected
	if c.connected_.IsSet() {
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

// connect

func (c *client) connect() (async.Future[internalConn], status.Status) {
	routine, ok := c.connecting.Unwrap()
	if ok {
		return routine, status.OK
	}

	routine = async.Run(c.connect1)
	c.connecting.Set(routine)
	return routine, status.OK
}

func (c *client) connect1(ctx async.Context) (internalConn, status.Status) {
	// Try to connect
	conn, st := c.connectRecover(ctx)

	// Clear connecting
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connecting.Unset()

	// Return if connected
	if st.OK() {
		return conn, st
	}

	// Return if cancelled/closed
	select {
	case <-ctx.Wait():
		return nil, ctx.Status()
	case <-c.closed_.Wait():
		return nil, status.Closedf("mpx client closed")
	default:
	}

	// Return error in on-demand mode
	if c.mode != ClientMode_AutoConnect {
		return nil, st
	}

	// Reconnect in auto-connect mode
	routine := async.Run(c.connect1)
	c.connecting.Set(routine)
	return nil, st
}

func (c *client) connectRecover(ctx async.Context) (_ internalConn, st status.Status) {
	defer func() {
		if e := recover(); e != nil {
			st = status.Recover(e)
		}
	}()

	// Increment attempt
	attempt := func() int {
		c.mu.Lock()
		defer c.mu.Unlock()

		c.connectAttempt++
		return c.connectAttempt
	}()

	// Sleep before reconnecting
	if attempt > 1 {
		timeout := reconnectTimeout(attempt)

		select {
		case <-ctx.Wait():
			return nil, ctx.Status()
		case <-time.After(timeout):
		}
	}

	// Connect
	conn, st := c.connector.connect(ctx, c.addr)
	if !st.OK() {
		return nil, st
	}
	go c.handle(conn)

	// Add connection
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed_.IsSet() {
		conn.Close()
		return nil, status.Closedf("mpx client closed")
	}

	c.conns = append(c.conns, conn)
	c.connectAttempt = 0

	c.connected_.Set()
	c.disconnected_.Unset()
	return conn, status.OK
}

func (c *client) handle(conn internalConn) {
	defer func() {
		if e := recover(); e != nil {
			st := status.Recover(e)
			c.logger.ErrorStatus("Connection panic", st)
		}
	}()

	st := conn.run()
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeClosed,
		status.CodeEnd:
	default:
		c.logger.ErrorStatus("Connection error", st)
	}
}

// util

// reconnectTimeout returns an exponential backoff timeout for reconnecting.
func reconnectTimeout(attempt int) time.Duration {
	multi := uint16(1<<attempt - 2)
	timeout := minConnectRetryTimeout * time.Duration(multi)
	return min(timeout, maxConnectRetryTimeout)
}
