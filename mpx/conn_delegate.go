// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

type connDelegate interface {
	// onConnClosed is called when the connection is closed.
	onConnClosed(c internalConn)

	// onConnChannelsReached is called when the number of channels reaches the target.
	onConnChannelsReached(c internalConn)
}

// noop

var _ connDelegate = noopConnDelegate{}

type noopConnDelegate struct{}

// onConnClosed is called when the connection is closed.
func (d noopConnDelegate) onConnClosed(c internalConn) {}

// onConnChannelsReached is called when the number of channels reaches the target.
func (d noopConnDelegate) onConnChannelsReached(c internalConn) {}
