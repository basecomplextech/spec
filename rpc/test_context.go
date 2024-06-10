package rpc

import "github.com/basecomplextech/baselibrary/async"

// TestContext returns a test context with a test connection.
func TestContext() Context {
	return newTestContext(async.NoContext())
}

// TestNextContext returns a test context with a test connection.
func TestNextContext(super async.Context) Context {
	return newTestContext(super)
}

// internal

var _ Context = (*testContext)(nil)

type testContext struct {
	async.Context
	disconnected async.Flag
}

func newTestContext(super async.Context) *testContext {
	return &testContext{
		Context:      super,
		disconnected: async.UnsetFlag(),
	}
}

// Disconnected returns a connection disconnected flag.
func (x *testContext) Disconnected() async.Flag {
	return x.disconnected
}

// AddConnListener adds a connection listener.
func (x *testContext) AddConnListener(ConnListener) (unsub func()) {
	return func() {}
}
