// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"sync"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/opt"
	"github.com/basecomplextech/baselibrary/status"
)

var _ connector = (*manualConnector)(nil)

// manualConnector manually connects on request.
type manualConnector struct {
	addr   string
	logger logging.Logger
	opts   Options

	closed_       async.MutFlag
	connected_    async.MutFlag
	disconnected_ async.MutFlag

	mu         sync.Mutex
	pool       connPool
	connecting opt.Opt[async.Routine[*conn]]
}

func newManualConnector(addr string, logger logging.Logger, opts Options) *manualConnector {
	return &manualConnector{
		addr:   addr,
		logger: logger,
		opts:   opts,

		closed_:       async.SetFlag(),
		connected_:    async.UnsetFlag(),
		disconnected_: async.SetFlag(),

		pool: newConnPool(),
	}
}

// closed returns a flag which indicates the connector is closed.
func (c *manualConnector) closed() async.Flag {
	return c.closed_
}

// connected returns a flag when there is at least one connected connection.
func (c *manualConnector) connected() async.Flag {
	return c.connected_
}

// disconnected returns a flag when there are no connected connections.
func (c *manualConnector) disconnected() async.Flag {
	return c.disconnected_
}

// methods

// connect returns a connection or a future.
func (c *manualConnector) conn(ctx async.Context) (*conn, async.Future[*conn], status.Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check closed
	if c.closed_.Get() {
		return nil, nil, status.Closedf("mpx client closed")
	}

	// Get connection with least number of channels
	// Remove disconnected connections.
	conn, ok := c.pool.conn()
	if ok {
		// Maybe connect more
		if len(c.pool) < c.opts.MaxConns {
			num := conn.channelNum()
			if num >= c.opts.TargetConnChannels {
				c.connect()
			}
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

// close stops and closes the connector.
func (c *manualConnector) close() status.Status {
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
	for conn := range c.pool {
		conn.Close()
	}
	clear(c.pool)

	// Update flags
	c.connected_.Unset()
	c.disconnected_.Set()
	return status.OK
}

// private

func (c *manualConnector) connect() (async.Future[*conn], status.Status) {
	routine, ok := c.connecting.Unwrap()
	if ok {
		return routine, status.OK
	}

	routine = async.Run(c.doConnect)
	c.connecting.Set(routine)
	return routine, status.OK
}

func (c *manualConnector) doConnect(ctx async.Context) (*conn, status.Status) {
	// Clear routine on exit
	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		c.connecting.Unset()
	}()

	// Connect
	conn, st := connect(c.addr, c.logger, c.opts)
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

	c.pool[conn] = struct{}{}
	c.connected_.Set()
	c.disconnected_.Unset()
	return conn, status.OK
}
