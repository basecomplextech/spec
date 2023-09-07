package rpc

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/prpc"
	"github.com/basecomplextech/spec/tcp"
)

// Client is a SpecRPC client.
type Client interface {
	// Close closes the client.
	Close() status.Status

	// Channel opens a new channel.
	Channel(cancel <-chan struct{}) (Channel, status.Status)

	// Request sends a request and returns status and result if status is OK.
	Request(cancel <-chan struct{}, req prpc.Request) (*alloc.Buffer, status.Status)
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

// Channel opens a new channel.
func (c *client) Channel(cancel <-chan struct{}) (Channel, status.Status) {
	tc, st := c.client.Connect(cancel)
	if !st.OK() {
		return nil, st
	}

	ok := false
	defer func() {
		if !ok {
			tc.Close()
		}
	}()

	tch, st := tc.Channel(cancel)
	if !st.OK() {
		return nil, st
	}
	defer func() {
		if !ok {
			tch.Close()
		}
	}()

	ch := newChannel(tc, tch)
	ok = true
	return ch, status.OK
}

// Request sends a request and returns status and result if status is OK.
func (c *client) Request(cancel <-chan struct{}, req prpc.Request) (*alloc.Buffer, status.Status) {
	ch, st := c.Channel(cancel)
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
