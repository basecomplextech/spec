package rpc

import (
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/prpc"
	"github.com/basecomplextech/spec/tcp"
)

// Conn is a single RPC connection.
type Conn interface {
	// Close closes the connection.
	Close() status.Status

	// Request sends a request and returns a response.
	Request(cancel <-chan struct{}, req *Request) (*ref.Box[prpc.Response], status.Status)

	// Internal

	// Free closes the connection and releases its resources.
	Free()
}

// Connect dials the given address and returns a connection.
func Connect(address string, logger logging.Logger) (Conn, status.Status) {
	c, st := tcp.Connect(address, logger)
	if !st.OK() {
		return nil, st
	}

	conn := newConn(c, logger)
	return conn, status.OK
}

// internal

type conn struct {
	conn   tcp.Conn
	logger logging.Logger
}

func newConn(c tcp.Conn, logger logging.Logger) Conn {
	return &conn{
		conn:   c,
		logger: logger,
	}
}

// Close closes the connection.
func (c *conn) Close() status.Status {
	return c.conn.Close()
}

// Request sends a request and returns a response.
func (c *conn) Request(cancel <-chan struct{}, req *Request) (*ref.Box[prpc.Response], status.Status) {
	// Build request
	preq, st := req.build()
	if !st.OK() {
		return nil, st
	}

	// Open stream
	stream, st := c.conn.Stream(cancel)
	if !st.OK() {
		return nil, st
	}

	// Free stream on error
	ok := false
	defer func() {
		if ok {
			return
		}
		stream.Free()
	}()

	// Write request
	if st := stream.Write(cancel, preq.Unwrap().Raw()); !st.OK() {
		return nil, st
	}

	// Read response
	msg, st := stream.Read(cancel)
	if !st.OK() {
		return nil, st
	}

	// Close stream
	if st := stream.Close(); !st.OK() {
		return nil, st
	}

	// Parse response
	presp, _, err := prpc.ParseResponse(msg)
	if err != nil {
		return nil, WrapError(err)
	}

	// Box response and stream
	box := ref.NewBox(presp, stream)
	ok = true
	return box, status.OK
}

// Internal

// Free closes the connection and releases its resources.
func (c *conn) Free() {
	c.conn.Free()
}