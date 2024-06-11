package rpc

import (
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/spec/mpx"
)

type TestContext = mpx.TestContext

// TestNewContext returns a test context with a test connection.
func TestNewContext() TestContext {
	return mpx.TestNewContext()
}

// TestNextContext returns a test context with a test connection.
func TestNextContext(super async.Context) TestContext {
	return mpx.TestNextContext(super)
}
