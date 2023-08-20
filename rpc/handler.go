package rpc

import (
	"io"
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

// HandlerFunc is a type adapter to allow the use of ordinary functions as RPC handlers.
type HandlerFunc func(cancel <-chan struct{}, req *ServerRequest, resp ServerResponse) status.Status

// Handle handles a server RPC request.
func (f HandlerFunc) Handle(cancel <-chan struct{}, req *ServerRequest, resp ServerResponse) status.Status {
	return f(cancel, req, resp)
}

// Request

// ServerRequest is a server RPC request.
type ServerRequest struct {
	buf *alloc.Buffer
	req prpc.Request
}

func newServerRequest(req *http.Request) (*ServerRequest, status.Status) {
	clen := req.ContentLength
	if clen < 0 {
		return nil, Error("Absent content length")
	}
	buf := alloc.NewBufferSize(int(clen))

	ok := false
	defer func() {
		if !ok {
			buf.Free()
		}
	}()

	b := buf.Grow(int(clen))
	if _, err := io.ReadFull(req.Body, b); err != nil {
		return nil, Errorf("Failed to read body: %v", err)
	}

	preq, _, err := prpc.ParseRequest(b)
	if err != nil {
		return nil, Errorf("Failed to parse request: %v", err)
	}

	req1 := &ServerRequest{
		buf: buf,
		req: preq,
	}
	ok = true
	return req1, status.OK
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

// free frees the request.
func (r *ServerRequest) free() {
	r.req = prpc.Request{}

	r.buf.Free()
	r.buf = nil
}

// Response

// ServerResponse writes the result of a server RPC call.
type ServerResponse interface {
	// Result adds a new result to the response.
	Result() prpc.ResultWriter
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
func (r *serverResponse) Result() prpc.ResultWriter {
	return r.results.Add()
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

	if err := r.results.End(); err != nil {
		return prpc.Response{}, status.WrapError(err)
	}

	p, err := r.resp.Build()
	if err != nil {
		return prpc.Response{}, status.WrapError(err)
	}
	return p, status.OK
}
