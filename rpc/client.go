package rpc

import (
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/prpc"
	"github.com/basecomplextech/spec/tcp"
)

// Client is a SpecRPC client.
type Client interface {
	// Close closes the client.
	Close() status.Status

	// Channel opens a channels and sends a request.
	Channel(cancel <-chan struct{}, req prpc.Request) (Channel, status.Status)

	// Request sends a request and returns a response.
	Request(cancel <-chan struct{}, req prpc.Request) (*ref.R[spec.Value], status.Status)
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

// Channel opens a channels and sends a request.
func (c *client) Channel(cancel <-chan struct{}, req prpc.Request) (Channel, status.Status) {
	ch, st := c.channel(cancel)
	if !st.OK() {
		return nil, st
	}

	ok := false
	defer func() {
		if !ok {
			ch.Free()
		}
	}()

	st = ch.Request(cancel, req)
	if !st.OK() {
		return nil, st
	}

	ok = true
	return ch, status.OK
}

// Request sends a request and returns a response.
func (c *client) Request(cancel <-chan struct{}, req prpc.Request) (*ref.R[spec.Value], status.Status) {
	ch, st := c.channel(cancel)
	if !st.OK() {
		return nil, st
	}
	defer ch.Free()

	st = ch.Request(cancel, req)
	if !st.OK() {
		return nil, st
	}

	return ch.Response(cancel)
}

// private

func (c *client) channel(cancel <-chan struct{}) (*channel, status.Status) {
	tch, st := c.client.Channel(cancel)
	if !st.OK() {
		return nil, st
	}

	ok := false
	defer func() {
		if !ok {
			tch.Close()
		}
	}()

	ch := newChannel(tch)
	ok = true
	return ch, status.OK
}
