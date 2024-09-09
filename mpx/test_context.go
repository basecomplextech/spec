// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import "github.com/basecomplextech/baselibrary/async"

type TestContext interface {
	Context

	// Disconnect sets the disconnected flag and calls the disconnected listeners.
	Disconnect()

	// OnDisconnectedNum returns the number of disconnect listeners.
	OnDisconnectedNum() int
}

// TestNewContext returns a test context with a test connection.
func TestNewContext() TestContext {
	return newTestContext(async.NoContext())
}

// TestNextContext returns a test context with a test connection.
func TestNextContext(super async.Context) TestContext {
	return newTestContext(super)
}

// internal

var _ Context = (*testContext)(nil)

type testContext struct {
	async.Context

	disconnected        async.MutFlag
	disconnectSeq       int
	disconnectListeners map[int]func()
}

func newTestContext(super async.Context) *testContext {
	return &testContext{
		Context: super,

		disconnected:        async.UnsetFlag(),
		disconnectListeners: make(map[int]func()),
	}
}

// Disconnect sets the disconnected flag and calls the disconnected listeners.
func (x *testContext) Disconnect() {
	x.disconnected.Set()

	for _, fn := range x.disconnectListeners {
		fn()
	}
}

// Disconnected returns a connection disconnected flag.
func (x *testContext) Disconnected() async.Flag {
	return x.disconnected
}

// OnDisconnected adds a disconnect listener, and returns an unsubscribe function.
func (x *testContext) OnDisconnected(fn func()) (unsub func()) {
	id := x.disconnectSeq
	x.disconnectSeq++
	x.disconnectListeners[id] = fn

	return func() {
		delete(x.disconnectListeners, id)
	}
}

// OnDisconnectedNum returns the number of disconnect listeners.
func (x *testContext) OnDisconnectedNum() int {
	return len(x.disconnectListeners)
}
