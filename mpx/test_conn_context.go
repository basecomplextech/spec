// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import "github.com/basecomplextech/baselibrary/async"

type TestConnContext interface {
	ConnContext

	// Disconnect sets the disconnected flag and calls the disconnected listeners.
	Disconnect()

	// OnDisconnectedNum returns the number of disconnect listeners.
	OnDisconnectedNum() int
}

// internal

var _ ConnContext = (*testConnContext)(nil)

type testConnContext struct {
	async.Context

	disconnected        async.MutFlag
	disconnectSeq       int
	disconnectListeners map[int]func()
}

func newTestConnContext(super async.Context) *testConnContext {
	return &testConnContext{
		Context: super,

		disconnected:        async.UnsetFlag(),
		disconnectListeners: make(map[int]func()),
	}
}

// Disconnect sets the disconnected flag and calls the disconnected listeners.
func (x *testConnContext) Disconnect() {
	x.disconnected.Set()

	for _, fn := range x.disconnectListeners {
		fn()
	}
}

// Disconnected returns a connection disconnected flag.
func (x *testConnContext) Disconnected() async.Flag {
	return x.disconnected
}

// OnDisconnected adds a disconnect listener, and returns an unsubscribe function.
func (x *testConnContext) OnDisconnected(fn func()) (unsub func()) {
	id := x.disconnectSeq
	x.disconnectSeq++
	x.disconnectListeners[id] = fn

	return func() {
		delete(x.disconnectListeners, id)
	}
}

// OnDisconnectedNum returns the number of disconnect listeners.
func (x *testConnContext) OnDisconnectedNum() int {
	return len(x.disconnectListeners)
}
