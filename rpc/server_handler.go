package rpc

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/prpc"
)

// Handler is an RPC handler.
type Handler interface {
	// Handle handles a request and returns its result and status.
	// Request is valid until the next call to any receive method.
	// Result is ignored if status is not OK.
	Handle(cancel <-chan struct{}, ch ServerChannel, req prpc.Request) (*alloc.Buffer, status.Status)
}

// HandleFunc is a type adapter to allow use of ordinary functions as RPC handlers.
type HandleFunc func(cancel <-chan struct{}, ch ServerChannel, req prpc.Request) (*alloc.Buffer, status.Status)

// Handle handles a request and returns its result and status.
func (f HandleFunc) Handle(cancel <-chan struct{}, ch ServerChannel, req prpc.Request) (*alloc.Buffer, status.Status) {
	return f(cancel, ch, req)
}
