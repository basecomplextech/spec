package rpc

import (
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/prpc"
)

// Clien is an RPC client.
type Client interface {
	// Free releases the client and its underlying transport.
	Free()

	// Request sends a request and returns a response.
	Request(cancel <-chan struct{}, req *Request) (prpc.Response, status.Status)
}

// NewClient returns a new client.
func NewClient() Client {
	return nil
}

// internal

var _ Client = (*client)(nil)

type client struct {
	transport Transport
}

func newClient(t Transport) *client {
	return &client{
		transport: t,
	}
}

// Free releases the transport.
func (c *client) Free() {
	c.transport.Free()
}

// Request sends a request and returns a response.
func (c *client) Request(cancel <-chan struct{}, req *Request) (prpc.Response, status.Status) {
	// Build request
	preq, st := req.Build()
	if !st.OK() {
		return prpc.Response{}, st
	}

	// Open stream
	stream, st := c.transport.Open(cancel)
	if !st.OK() {
		return prpc.Response{}, st
	}
	defer stream.Free()

	// Send request
	breq := preq.Unwrap().Raw()
	if st := stream.Send(cancel, breq); !st.OK() {
		return prpc.Response{}, st
	}

	// Receive response
	bresp, st := stream.Receive(cancel)
	if !st.OK() {
		return prpc.Response{}, st
	}

	// Parse response
	resp, _, err := prpc.ParseResponse(bresp)
	if err != nil {
		return prpc.Response{}, status.WrapError(err)
	}

	// Return response
	return resp, status.OK
}
