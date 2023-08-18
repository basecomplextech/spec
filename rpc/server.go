package rpc

import (
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/prpc"
)

// Handler is a server RPC handler.
type Handler interface {
	// Handle handles a server RPC request.
	Handle(cancel <-chan struct{}, req *ServerRequest, resp ServerResponse) status.Status
}

// ServerRequest is a server RPC request.
type ServerRequest struct {
	req prpc.Request
}

// ServerResponse writes the result of a server RPC call.
type ServerResponse interface {
	// Result adds a new result to the response.
	Result(name string) spec.FieldWriter
}

// Call returns an RPC call by index or panics on out of range.
func (r *ServerRequest) Call(i int) (prpc.Call, status.Status) {
	n := r.req.Calls().Len()
	if i < 0 || i >= n {
		return prpc.Call{}, status.Newf("rpc_error", "Call index %d out of range [0, %d)", i, n)
	}

	call, err := r.req.Calls().GetErr(i)
	if err != nil {
		return prpc.Call{}, status.Newf("rpc_error", "Failed to parse call %d: %v", i, err)
	}
	return call, status.OK
}

// Calls returns the number of RPC calls.
func (r *ServerRequest) Calls() int {
	return r.req.Calls().Len()
}
