package rpc

import (
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
)

// Handler is an RPC handler.
type Handler interface {
	// Handle handles a request and returns its result and status.
	// Result is ignored if status is not OK.
	Handle(ctx async.Context, ch ServerChannel) (*ref.R[[]byte], status.Status)
}

// Subhandler is an RPC subservice handler.
type Subhandler interface {
	// Handle handles a request and returns its result and status.
	// Result is ignored if status is not OK.
	Handle(ctx async.Context, ch ServerChannel, index int) (*ref.R[[]byte], status.Status)
}

// HandleFunc is a type adapter to allow use of ordinary functions as RPC handlers.
type HandleFunc func(ctx async.Context, ch ServerChannel) (*ref.R[[]byte], status.Status)

// Handle handles a request and returns its result and status.
// Result is ignored if status is not OK.
func (f HandleFunc) Handle(ctx async.Context, ch ServerChannel) (*ref.R[[]byte], status.Status) {
	return f(ctx, ch)
}
