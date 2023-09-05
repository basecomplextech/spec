package rpc

import (
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/prpc"
)

// Handler is an RCP handler.
type Handler interface {
	// Handle handles a request and returns a response.
	Handle(cancel <-chan struct{}, req prpc.Request) (*Response, status.Status)
}

// HandleFunc is a type adapter to allow use of ordinary functions as RPC handlers.
type HandleFunc func(cancel <-chan struct{}, req prpc.Request) (*Response, status.Status)

// Handle handles a request and returns a response.
func (f HandleFunc) Handle(cancel <-chan struct{}, req prpc.Request) (*Response, status.Status) {
	return f(cancel, req)
}
