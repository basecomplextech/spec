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

	// OnDisconnected adds a disconnect listener, and returns an unsubscribe function.
	//
	// The unsubscribe function does not deadlock, even if the listener is being called right now.
	OnDisconnected(fn func()) (unsub func())
}

// internal

var _ ConnContext = (*connContext)(nil)

type connContext struct {
	async.Context
	conn internalConn
}

func newConnContext(conn internalConn) *connContext {
	return &connContext{
		Context: async.NewContext(),
		conn:    conn,
	}
}

// Disconnected returns a connection disconnected flag.
func (c *connContext) Disconnected() async.Flag {
	if c.conn == nil {
		return async.UnsetFlag()
	}
	return c.conn.Closed()
}

// OnDisconnected adds a disconnect listener, and returns an unsubscribe function.
func (c *connContext) OnDisconnected(fn func()) (unsub func()) {
	if c.conn == nil {
		fn()
		return func() {}
	}

	return c.conn.OnClosed(fn)
}
