// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

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
