package rpc

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/prpc"
)

// Response is an outgoing RPC response.
type Response struct {
	buf     *alloc.Buffer
	resp    prpc.ResponseWriter
	results spec.MessageListWriter[prpc.ResultWriter]

	done  bool
	freed bool
}

// NewResponse returns a new response.
func NewResponse() *Response {
	buf := acquireBuffer()
	resp := prpc.NewResponseWriterBuffer(buf)
	results := resp.Results()

	return &Response{
		buf:     buf,
		resp:    resp,
		results: results,
	}
}

// Free releases the response resources.
func (r *Response) Free() {
	if r.freed {
		return
	}

	r.done = true
	r.freed = true

	r.results = spec.MessageListWriter[prpc.ResultWriter]{}
	r.resp = prpc.ResponseWriter{}

	releaseBuffer(r.buf)
	r.buf = nil
}

// Add adds a result to the response and returns a result writer.
func (r *Response) Add() prpc.ResultWriter {
	if r.done {
		panic("response is done")
	}

	return r.results.Add()
}

// internal

// build builds and returns the response data, data is valid until the response is freed.
func (r *Response) build() (prpc.Response, status.Status) {
	if r.done {
		panic("response is done")
	}
	r.done = true

	if err := r.results.End(); err != nil {
		return prpc.Response{}, WrapError(err)
	}

	p, err := r.resp.Build()
	if err != nil {
		return prpc.Response{}, WrapError(err)
	}
	return p, status.OK
}
