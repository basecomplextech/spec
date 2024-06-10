package mpx

import "github.com/basecomplextech/baselibrary/async"

type Context interface {
	async.Context

	// Disconnected returns a connection disconnected flag.
	Disconnected() async.Flag

	// AddConnListener adds a connection listener.
	AddConnListener(ConnListener) (unsub func())
}

// TestContext returns a test context without a connection.

// ClosedContext returns a closed context.
func ClosedContext() Context {
	return closedContext
}

// internal

var _ Context = (*context)(nil)

type context struct {
	async.Context
	conn internalConn
}

var closedContext = func() *context {
	return &context{
		Context: async.CancelledContext(),
		conn:    nil,
	}
}()

func newContext(conn internalConn) *context {
	return &context{
		Context: async.NewContext(),
		conn:    conn,
	}
}

// Disconnected returns a connection disconnected flag.
func (c *context) Disconnected() async.Flag {
	if c.conn == nil {
		return async.UnsetFlag()
	}
	return c.conn.Closed()
}

// AddConnListener adds a connection listener.
func (c *context) AddConnListener(l ConnListener) (unsub func()) {
	if c.conn == nil {
		l.OnDisconnected(nil)
		return func() {}
	}

	return c.conn.AddListener(l)
}
