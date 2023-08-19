package rpc

import (
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/prpc"
)

// Clien is an RPC client.
type Client interface {
	// Free releases the client and its underlying connector.
	Free()

	// Request sends a request and returns a response.
	Request(cancel <-chan struct{}, req *Request) (prpc.Response, status.Status)
}

// NewClient returns a new client.
func NewClient(c Connector) Client {
	return newClient(c)
}

// internal

var _ Client = (*client)(nil)

type client struct {
	connector Connector
}

func newClient(c Connector) *client {
	return &client{
		connector: c,
	}
}

// Free releases the client.
func (c *client) Free() {
	c.connector.Free()
}

// Request sends a request and returns a response.
func (c *client) Request(cancel <-chan struct{}, req *Request) (prpc.Response, status.Status) {
	// Build request
	preq, st := req.Build()
	if !st.OK() {
		return prpc.Response{}, st
	}

	// Open connection
	conn, st := c.connector.Connect(cancel)
	if !st.OK() {
		return prpc.Response{}, st
	}
	defer conn.Free()

	// Send request
	breq := preq.Unwrap().Raw()
	if st := conn.Send(cancel, breq); !st.OK() {
		return prpc.Response{}, st
	}

	// Receive response
	bresp, st := conn.Receive(cancel)
	if !st.OK() {
		return prpc.Response{}, st
	}

	// Parse response
	presp, _, err := prpc.ParseResponse(bresp)
	if err != nil {
		return prpc.Response{}, status.WrapError(err)
	}

	// Return response
	return presp, status.OK
}
