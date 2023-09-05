package rpc

import (
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/tcp"
)

// Client is a SpecRPC client.
type Client interface {
	// Close closes the client.
	Close() status.Status

	// Connect connects to an address.
	Connect(cancel <-chan struct{}) (Conn, status.Status)
}

// internal

var _ Client = (*client)(nil)

type client struct {
	client tcp.Client
	logger logging.Logger
}

func newClient(address string, logger logging.Logger) *client {
	return &client{
		client: tcp.NewClient(address, logger),
		logger: logger,
	}
}

// Close closes the client.
func (c *client) Close() status.Status {
	return c.client.Close()
}

// Connect connects to an address.
func (c *client) Connect(cancel <-chan struct{}) (Conn, status.Status) {
	tc, st := c.client.Connect(cancel)
	if !st.OK() {
		return nil, st
	}

	conn := newConn(tc, c.logger)
	return conn, status.OK
}
