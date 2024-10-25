// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

type connDelegate interface {
	// onConnClosed is called when the connection is closed.
	onConnClosed(c conn)

	// onConnChannelsReached is called when the number of channels reaches the target.
	// The method is used by the auto connector to establish more connections.
	onConnChannelsReached(c conn)
}

// noop

var _ connDelegate = noopConnDelegate{}

type noopConnDelegate struct{}

// onConnClosed is called when the connection is closed.
func (d noopConnDelegate) onConnClosed(c conn) {}

// onConnChannelsReached is called when the number of channels reaches the target.
// The method is used by the auto connector to establish more connections.
func (d noopConnDelegate) onConnChannelsReached(c conn) {}
