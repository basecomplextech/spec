// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import "github.com/basecomplextech/baselibrary/async"

// Context is a channel context.
type Context interface {
	async.Context

	// Conn returns a connection context.
	Conn() ConnContext
}

// TestContext returns a test context without a connection.

// ClosedContext returns a closed context.
func ClosedContext() Context {
	return closedContext
}

// internal

var _ Context = (*context)(nil)

type context struct {
	async.CancelContext
	conn internalConn
}

var closedContext = func() *context {
	return &context{
		CancelContext: async.CancelledContext(),
		conn:          nil,
	}
}()

func newContext(conn internalConn) *context {
	return &context{
		CancelContext: async.NewContext(),
		conn:          conn,
	}
}

// Conn returns a connection context.
func (c *context) Conn() ConnContext {
	return c.conn.Context()
}
