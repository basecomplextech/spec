package rpc

import (
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/prpc"
	"github.com/basecomplextech/spec/tcp"
)

// Client is a SpecRPC client.
type Client interface {
	// Close closes the client.
	Close() status.Status

	// Request sends a request and returns a channel.
	Request(cancel <-chan struct{}, req prpc.Request) (Channel, status.Status)
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

// Request sends a request and returns a channel.
func (c *client) Request(cancel <-chan struct{}, req prpc.Request) (Channel, status.Status) {
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
