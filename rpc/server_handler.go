// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package rpc

import (
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
)

// Handler is an RPC handler.
type Handler interface {
	// Handle handles a request and returns its result and status.
	// Result is ignored if status is not OK.
	Handle(ctx Context, ch ServerChannel) (ref.R[[]byte], status.Status)
}

// Subhandler is an RPC subservice handler.
type Subhandler interface {
	// Handle handles a request and returns its result and status.
	// Result is ignored if status is not OK.
	Handle(ctx Context, ch ServerChannel, index int) (ref.R[[]byte], status.Status)
}

// NextHandler is an RPC next call handler in a call chain.
type NextHandler[T any] interface {
	Handle(T) status.Status
}

type Subhandler1[T any] interface {
	NextHandler[T]

	Result() ref.R[[]byte]
	Free()
}

// HandleFunc is a type adapter to allow use of ordinary functions as RPC handlers.
type HandleFunc func(ctx Context, ch ServerChannel) (ref.R[[]byte], status.Status)

// Handle handles a request and returns its result and status.
// Result is ignored if status is not OK.
func (f HandleFunc) Handle(ctx Context, ch ServerChannel) (ref.R[[]byte], status.Status) {
	return f(ctx, ch)
}
