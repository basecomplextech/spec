// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package rpc

import (
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/mpx"
	"github.com/basecomplextech/spec/proto/prpc"
)

// ClientMode specifies how the client connects to the server.
type ClientMode = mpx.ClientMode

const (
	// ClientMode_OnDemand connects to the server on demand, does not reconnect on errors.
	ClientMode_OnDemand = mpx.ClientMode_OnDemand

	// ClientMode_AutoConnect automatically connects and reconnects to the server.
	// The client reconnects with exponential backoff on errors.
	ClientMode_AutoConnec = mpx.ClientMode_AutoConnect
)

// Client is a SpecRPC client.
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

	// Close closes the client, disconnect from theh server.
	Close() status.Status

	// Methods

	// Channel opens a channels and sends a request.
	Channel(ctx async.Context, req prpc.Request) (Channel, status.Status)

	// Request sends a request and returns a response.
	Request(ctx async.Context, req prpc.Request) (ref.R[spec.Value], status.Status)

	// RequestOneway sends a request and closes the channel, without waiting for a response.
	RequestOneway(ctx async.Context, req prpc.Request) status.Status

	// Internal

	// Unwrap returns the internal client.
	Unwrap() mpx.Client
}

// NewClient returns a new client.
func NewClient(address string, mode ClientMode, logger logging.Logger, opts Options) Client {
	return newClient(address, mode, logger, opts)
}

// internal

var _ Client = (*client)(nil)

type client struct {
	client mpx.Client
	logger logging.Logger
}

func newClient(address string, mode ClientMode, logger logging.Logger, opts Options) *client {
	return &client{
		client: mpx.NewClient(address, mode, logger, opts),
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

// Flags

// Closed indicates that the client is closed.
func (c *client) Closed() async.Flag {
	return c.client.Closed()
}

// Connected indicates that the client is connected to the server.
func (c *client) Connected() async.Flag {
	return c.client.Connected()
}

// Disconnected indicates that the client is disconnected from the server.
func (c *client) Disconnected() async.Flag {
	return c.client.Disconnected()
}

// Lifecycle

// Close closes the client.
func (c *client) Close() status.Status {
	return c.client.Close()
}

// Methods

// Channel opens a channels and sends a request.
func (c *client) Channel(ctx async.Context, req prpc.Request) (Channel, status.Status) {
	// Open channel
	ch, st := c.channel(ctx)
	switch st.Code {
	case status.CodeOK:
	case status.CodeCancelled:
		return nil, st
	default:
		c.logger.ErrorStatus("RPC client request error", st)
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
		method := ch.Method()
		c.logger.ErrorStatus("RPC client request error", st, "method", method)
		return nil, st
	}

	// Done
	ok = true
	return ch, status.OK
}

// Request sends a request and returns a response.
func (c *client) Request(ctx async.Context, req prpc.Request) (ref.R[spec.Value], status.Status) {
	// Open channel
	ch, st := c.channel(ctx)
	switch st.Code {
	case status.CodeOK:
	case status.CodeCancelled:
		return nil, st
	default:
		c.logger.ErrorStatus("RPC client request error", st)
		return nil, st
	}

	// Free on error
	done := false
	defer func() {
		if !done {
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
		method := ch.Method()
		c.logger.ErrorStatus("RPC client request error", st, "method", method)
		return nil, st
	}

	// Read response
	result, st := ch.Response(ctx)
	if !st.OK() {
		return nil, st
	}

	result_ := ref.NewFreer(result, ch)
	done = true
	return result_, status.OK
}

// RequestOneway sends a request and closes the channel, without waiting for a response.
func (c *client) RequestOneway(ctx async.Context, req prpc.Request) status.Status {
	// Open channel
	ch, st := c.channel(ctx)
	switch st.Code {
	case status.CodeOK:
	case status.CodeCancelled:
		return st
	default:
		c.logger.ErrorStatus("RPC client request error", st)
		return st
	}
	defer ch.Free()

	// Send request
	st = ch.Request(ctx, req)
	switch st.Code {
	case status.CodeOK:
	case status.CodeCancelled:
		return st
	default:
		method := ch.Method()
		c.logger.ErrorStatus("RPC client request error", st, "method", method)
		return st
	}

	// Do not wait for response
	return status.OK
}

// Internal

// Unwrap returns the internal client.
func (c *client) Unwrap() mpx.Client {
	return c.client
}

// private

func (c *client) channel(ctx async.Context) (*channel, status.Status) {
	ch, st := c.client.Channel(ctx)
	if !st.OK() {
		return nil, st
	}

	ok := false
	defer func() {
		if !ok {
			ch.Free()
		}
	}()

	ch1 := newChannel(ch, c.logger)
	ok = true
	return ch1, status.OK
}
