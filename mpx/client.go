// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
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

	// Flags

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

	connector connector // manual or auto connector
}

func newClient(addr string, logger logging.Logger, opts Options) *client {
	return &client{
		addr:    addr,
		logger:  logger,
		options: opts.clean(),

		connector: newManualConnector(addr, logger, opts),
	}
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

// Connected indicates that the client is connected to the server.
func (c *client) Connected() async.Flag {
	return c.connector.connected()
}

// Disconnected indicates that the client is disconnected from the server.
func (c *client) Disconnected() async.Flag {
	return c.connector.disconnected()
}

// Lifecycle

// Close closes the client.
func (c *client) Close() status.Status {
	return c.connector.close()
}

// Methods

// Conn returns a connection.
func (c *client) Conn(ctx async.Context) (Conn, status.Status) {
	// Get connection
	conn, future, st := c.connector.conn(ctx)
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
