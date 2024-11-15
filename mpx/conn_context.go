// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import "github.com/basecomplextech/baselibrary/async"

// ConnContext is a connection context.
type ConnContext interface {
	async.Context

	// Disconnected returns a connection disconnected flag.
	Disconnected() async.Flag

	// OnDisconnected adds a disconnect listener, and returns an unsubscribe function,
	// or false if the connection is already closed.
	OnDisconnected(fn func()) (unsub func(), _ bool)
}

// internal

var _ ConnContext = (*connContext)(nil)

type connContext struct {
	async.CancelContext
	conn internalConn
}

func newConnContext(conn internalConn) *connContext {
	return &connContext{
		CancelContext: async.NewContext(),
		conn:          conn,
	}
}

// Disconnected returns a connection disconnected flag.
func (c *connContext) Disconnected() async.Flag {
	if c.conn == nil {
		return async.UnsetFlag()
	}
	return c.conn.Closed()
}

// OnDisconnected adds a disconnect listener, and returns an unsubscribe function,
// or false if the connection is already closed.
func (c *connContext) OnDisconnected(fn func()) (unsub func(), _ bool) {
	if c.conn == nil {
		return nil, false
	}

	return c.conn.OnClosed(fn)
}
