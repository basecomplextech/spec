package rpc

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/pools"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/prpc"
)

// Request is an outgoing RPC request builder.
type Request struct {
	s *requestState
}

// NewRequest returns a new request.
func NewRequest() *Request {
	s := acquireRequestState()
	return &Request{s: s}
}

// Free releases the request resources.
func (r *Request) Free() {
	if r.s == nil {
		return
	}

	s := r.s
	r.s = nil
	releaseRequestState(s)
}

// Add adds a call and returns its writer.
func (r *Request) Add(method string) prpc.CallWriter {
	s := r.state()
	if s.done {
		panic("request is done")
	}

	call := s.calls.Add()
	call.Method(method)
	return call
}

// AddEmpty adds a call with no input.
func (r *Request) AddEmpty(method string) status.Status {
	s := r.state()
	if s.done {
		panic("request is done")
	}

	call := s.calls.Add()
	call.Method(method)

	if err := call.End(); err != nil {
		return WrapError(err)
	}
	return status.OK
}

// AddMessage adds a call with an input message.
func (r *Request) AddMessage(method string, input spec.Message) status.Status {
	s := r.state()
	if s.done {
		panic("request is done")
	}

	call := s.calls.Add()
	call.Method(method)
	call.CopyInput(input)

	if err := call.End(); err != nil {
		return WrapError(err)
	}
	return status.OK
}

// Build builds and returns the request data, data is valid until the request is freed.
func (r *Request) Build() (prpc.Request, status.Status) {
	s := r.state()
	if s.done {
		panic("request is done")
	}

	if err := s.calls.End(); err != nil {
		return prpc.Request{}, WrapError(err)
	}

	p, err := s.req.Build()
	if err != nil {
		return prpc.Request{}, WrapError(err)
	}
	return p, status.OK
}

// internal

func (r *Request) state() *requestState {
	if r.s == nil {
		panic("request is freed")
	}
	return r.s
}

// state

var requestStatePool = pools.NewPoolFunc(newRequestState)

type requestState struct {
	buf    alloc.Buffer
	writer spec.Writer

	req   prpc.RequestWriter
	calls spec.MessageListWriter[prpc.CallWriter]
	done  bool
}

func acquireRequestState() *requestState {
	return requestStatePool.New()
}

func releaseRequestState(s *requestState) {
	s.reset()
	requestStatePool.Put(s)
}

func newRequestState() *requestState {
	buf := alloc.NewBuffer()
	writer := spec.NewWriterBuffer(buf)

	req := prpc.NewRequestWriterBuffer(buf)
	calls := req.Calls()

	return &requestState{
		buf:    buf,
		writer: writer,

		req:   req,
		calls: calls,
	}
}

func (s *requestState) reset() {
	s.buf.Reset()
	s.writer.Reset(s.buf)

	s.req = prpc.NewRequestWriterTo(s.writer.Message())
	s.calls = s.req.Calls()
	s.done = false
}
