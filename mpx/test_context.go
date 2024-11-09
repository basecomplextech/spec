// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import "github.com/basecomplextech/baselibrary/async"

type TestContext interface {
	Context
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
	conn *testConnContext
}

func newTestContext(super async.Context) *testContext {
	return &testContext{
		Context: super,
		conn:    newTestConnContext(super),
	}
}

// Conn returns a connection context.
func (x *testContext) Conn() ConnContext {
	return x.conn
}
