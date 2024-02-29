package rpc

import (
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/prpc"
	"github.com/basecomplextech/spec/tcp"
)

// Client is a SpecRPC client.
type Client interface {
	// Address returns the server address.
	Address() string

	// Options returns the client options.
	Options() Options

	// IsConnected returns true if the client is connected to the server.
	IsConnected() bool

	// Async

	// Changed adds a connected/disconnected listener.
	Changed() (<-chan struct{}, func())

	// Connected indicates that the client is connected to the server.
	Connected() <-chan struct{}

	// Disconnected indicates that the client is disconnected from the server.
	Disconnected() <-chan struct{}

	// Methods

	// Connect manually starts the internal connect loop.
	Connect() status.Status

	// Close closes the client, disconnect from theh server.
	Close() status.Status

	// Channel opens a channels and sends a request.
	Channel(ctx async.Context, req prpc.Request) (Channel, status.Status)

	// Request sends a request and returns a response.
	Request(ctx async.Context, req prpc.Request) (*ref.R[spec.Value], status.Status)

	// Internal

	// Unwrap returns the internal client.
	Unwrap() tcp.Client
}

// NewClient returns a new client.
func NewClient(address string, logger logging.Logger, opts Options) Client {
	return newClient(address, logger, opts)
}

// internal

var _ Client = (*client)(nil)

type client struct {
	client tcp.Client
	logger logging.Logger
}

func newClient(address string, logger logging.Logger, opts Options) *client {
	return &client{
		client: tcp.NewClient(address, logger, opts),
		logger: logger,
	}
}

// Address returns the server address.
func (c *client) Address() string {
	return c.client.Address()
}

// Options returns the client options.
func (c *client) Options() Options {
	return c.client.Options()
}

// IsConnected returns true if the client is connected to the server.
func (c *client) IsConnected() bool {
	return c.client.IsConnected()
}

// Async

// Changed adds a connected/disconnected listener.
func (c *client) Changed() (<-chan struct{}, func()) {
	return c.client.Changed()
}

// Connected indicates that the client is connected to the server.
func (c *client) Connected() <-chan struct{} {
	return c.client.Connected()
}

// Disconnected indicates that the client is disconnected from the server.
func (c *client) Disconnected() <-chan struct{} {
	return c.client.Disconnected()
}

// Methods

// Connect manually starts the internal connect loop.
func (c *client) Connect() status.Status {
	return c.client.Connect()
}

// Close closes the client.
func (c *client) Close() status.Status {
	return c.client.Close()
}

// Channel opens a channels and sends a request.
func (c *client) Channel(ctx async.Context, req prpc.Request) (Channel, status.Status) {
	// Open channel
	ch, st := c.channel(ctx)
	switch st.Code {
	case status.CodeOK:
	case status.CodeCancelled:
		return nil, st
	default:
		method := requestMethod(req)
		c.logger.ErrorStatus("RPC client request error", st, "method", method)
		return nil, st
	}

	// Free on error
	ok := false
	defer func() {
		if !ok {
			ch.Free()
		}
	}()

	// Send request
	st = ch.Request(ctx, req)
	switch st.Code {
	case status.CodeOK:
	case status.CodeCancelled:
		return nil, st
	default:
		method := requestMethod(req)
		c.logger.ErrorStatus("RPC client request error", st, "method", method)
		return nil, st
	}

	// Done
	ok = true
	return ch, status.OK
}

// Request sends a request and returns a response.
func (c *client) Request(ctx async.Context, req prpc.Request) (*ref.R[spec.Value], status.Status) {
	// Open channel
	ch, st := c.channel(ctx)
	switch st.Code {
	case status.CodeOK:
	case status.CodeCancelled:
		return nil, st
	default:
		method := requestMethod(req)
		c.logger.ErrorStatus("RPC client request error", st, "method", method)
		return nil, st
	}
	defer ch.Free()

	// Send request
	st = ch.Request(ctx, req)
	switch st.Code {
	case status.CodeOK:
	case status.CodeCancelled:
		return nil, st
	default:
		method := requestMethod(req)
		c.logger.ErrorStatus("RPC client request error", st, "method", method)
		return nil, st
	}

	// Read response
	return ch.Response(ctx)
}

// Internal

// Unwrap returns the internal client.
func (c *client) Unwrap() tcp.Client {
	return c.client
}

// private

func (c *client) channel(ctx async.Context) (*channel, status.Status) {
	tch, st := c.client.Channel(ctx)
	if !st.OK() {
		return nil, st
	}

	ok := false
	defer func() {
		if !ok {
			tch.Close()
		}
	}()

	ch := newChannel(tch, c.logger)
	ok = true
	return ch, status.OK
}
