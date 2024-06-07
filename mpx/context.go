package mpx

import "github.com/basecomplextech/baselibrary/async"

type Context interface {
	async.Context

	// Disconnected returns a connection disconnected flag.
	Disconnected() async.Flag
}

// internal

var _ Context = (*context)(nil)

type context struct {
	async.Context

	closed async.Flag
}

func newContext(conn *conn) *context {
	return &context{
		Context: async.NewContext(),
		closed:  conn.closed,
	}
}

// Disconnected returns a connection disconnected flag.
func (c *context) Disconnected() async.Flag {
	return c.closed
}
