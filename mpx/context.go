// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import "github.com/basecomplextech/baselibrary/async"

type Context interface {
	async.Context

	// Disconnected returns a connection disconnected flag.
	Disconnected() async.Flag

	// OnDisconnected adds a disconnect listener, and returns an unsubscribe function.
	//
	// The unsubscribe function does not deadlock, even if the listener is being called right now.
	OnDisconnected(fn func()) (unsub func())
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

// OnDisconnected adds a disconnect listener, and returns an unsubscribe function.
func (c *context) OnDisconnected(fn func()) (unsub func()) {
	if c.conn == nil {
		fn()
		return func() {}
	}

	return c.conn.OnClosed(fn)
}
