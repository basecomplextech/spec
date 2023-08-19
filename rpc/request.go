package rpc

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/prpc"
)

// Request is a client RPC request.
type Request struct {
	buf *alloc.Buffer

	req   prpc.RequestWriter
	calls spec.MessageListWriter[prpc.CallWriter]

	done  bool
	freed bool
}

func NewRequest() *Request {
	buf := alloc.NewBuffer()
	req := prpc.NewRequestWriterBuffer(buf)
	calls := req.Calls()

	return &Request{
		buf:   buf,
		req:   req,
		calls: calls,
	}
}

// Free releases the resources associated with the request.
func (r *Request) Free() {
	if r.freed {
		return
	}

	r.done = true
	r.freed = true

	r.calls = spec.MessageListWriter[prpc.CallWriter]{}
	r.req = prpc.RequestWriter{}

	r.buf.Free()
	r.buf = nil
}

// Call adds a call to the request and returns a call writer.
func (r *Request) Call(method string) prpc.CallWriter {
	if r.done {
		panic("request is done")
	}

	return r.calls.Add()
}

// Build builds and returns the request data, data is valid until the request is freed.
func (r *Request) Build() (prpc.Request, status.Status) {
	if r.done {
		return prpc.Request{}, status.Error("request is done")
	}
	r.done = true

	if err := r.calls.End(); err != nil {
		return prpc.Request{}, status.WrapError(err)
	}

	p, err := r.req.Build()
	if err != nil {
		return prpc.Request{}, status.WrapError(err)
	}
	return p, status.OK
}