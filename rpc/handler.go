package rpc

import (
	"net/http"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/prpc"
)

// Handler is a server RPC handler.
type Handler interface {
	// Handle handles a server RPC request.
	Handle(cancel <-chan struct{}, req *ServerRequest, resp ServerResponse) status.Status
}

// Request

// ServerRequest is a server RPC request.
type ServerRequest struct {
	req prpc.Request
}

func newServerRequest(req *http.Request) (*ServerRequest, status.Status) {
	return nil, status.OK
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

// Response

// ServerResponse writes the result of a server RPC call.
type ServerResponse interface {
	// Result adds a new result to the response.
	Result(name string) spec.FieldWriter
}

// internal

var _ ServerResponse = (*serverResponse)(nil)

type serverResponse struct {
	buf     *alloc.Buffer
	resp    prpc.ResponseWriter
	results spec.MessageListWriter[prpc.ResultWriter]

	done bool
}

func newServerResponse() *serverResponse {
	buf := alloc.NewBuffer()
	resp := prpc.NewResponseWriterBuffer(buf)
	results := resp.Results()

	return &serverResponse{
		buf:     buf,
		resp:    resp,
		results: results,
	}
}

// Result adds a new result to the response.
func (r *serverResponse) Result(name string) spec.FieldWriter {
	result := r.results.Add()
	result.Name(name)
	return result.Value()
}

func (r *serverResponse) free() {
	r.results = spec.MessageListWriter[prpc.ResultWriter]{}
	r.resp = prpc.ResponseWriter{}

	r.buf.Free()
	r.buf = nil
}

func (r *serverResponse) build() (prpc.Response, status.Status) {
	if r.done {
		return prpc.Response{}, status.New("rpc_error", "Response already completed")
	}
	r.done = true

	p, err := r.resp.Build()
	if err != nil {
		return prpc.Response{}, status.WrapError(err)
	}
	return p, status.OK
}
