package mpx

import "github.com/basecomplextech/baselibrary/async"

type Context interface {
	async.Context

	// Disconnected returns a connection disconnected flag.
	Disconnected() async.Flag
}

// ClosedContext returns a closed context.
func ClosedContext() Context {
	return closedContext
}

// internal

var _ Context = (*context)(nil)

type context struct {
	async.Context
	closed async.Flag
}

var closedContext = func() *context {
	return &context{
		Context: async.CancelledContext(),
		closed:  async.SetFlag(),
	}
}()

func newContext(conn internalConn) *context {
	return &context{
		Context: async.NewContext(),
		closed:  conn.Closed(),
	}
}

// Disconnected returns a connection disconnected flag.
func (c *context) Disconnected() async.Flag {
	return c.closed
}
