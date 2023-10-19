package rpc

import (
	"strings"
	"sync"
	"time"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/prpc"
	"github.com/basecomplextech/spec/tcp"
)

// Server is an RPC server.
type Server interface {
	async.Service

	// Address returns the address the server is listening to.
	Address() string

	// Listening indicates that the server is listening.
	Listening() <-chan struct{}

	// Options returns the server options.
	Options() Options
}

// NewServer returns a new RPC server.
func NewServer(address string, handler Handler, logger logging.Logger, opts Options) Server {
	return newServer(address, handler, logger, opts)
}

// internal

type server struct {
	tcp.Server

	handler Handler
	logger  logging.Logger
}

func newServer(address string, handler Handler, logger logging.Logger, opts Options) *server {
	s := &server{
		handler: handler,
		logger:  logger,
	}
	s.Server = tcp.NewServer(address, s, logger, opts)
	return s
}

// HandleChannel handles an incoming TCP channel.
func (s *server) HandleChannel(tch tcp.Channel) (st status.Status) {
	// Read message
	b, st := tch.Read(nil)
	if !st.OK() {
		return st
	}

	// Parse message
	msg, _, err := prpc.ParseMessage(b)
	if err != nil {
		return WrapErrorf(err, "failed to parse request message")
	}

	// Check request
	typ := msg.Type()
	if typ != prpc.MessageType_Request {
		return Errorf("unexpected request message type %d, expected %d",
			typ, prpc.MessageType_Request)
	}

	// Handle request
	result, st := s.handleRequest(tch, msg.Req())
	if result != nil {
		defer result.Release()
	}

	// Make response
	var resp []byte
	{
		buf := acquireBuffer()
		defer releaseBuffer(buf)

		w := prpc.NewMessageWriterBuffer(buf)
		w.Type(prpc.MessageType_Response)
		{
			w1 := w.Resp()
			w2 := w1.Status()
			w2.Code(string(st.Code))
			w2.Message(st.Message)
			if err := w2.End(); err != nil {
				return WrapError(err)
			}
			if result != nil {
				w1.Result().Any(result.Unwrap())
			}
			if err := w1.End(); err != nil {
				return WrapError(err)
			}
		}

		msg, err := w.Build()
		if err != nil {
			return WrapError(err)
		}
		resp = msg.Unwrap().Raw()
	}

	// Write response
	return tch.Write(nil, resp)
}

// private

func (s *server) handleRequest(tch tcp.Channel, req prpc.Request) (result *ref.R[[]byte], st status.Status) {
	start := time.Now()
	method := requestMethod(req)

	// Handle panic
	defer func() {
		e := recover()
		if e == nil {
			return
		}

		st = status.Recover(e)
		s.logger.ErrorStatus("Request panic", st, "method", method)
	}()

	// Handle request
	ch := newServerChannel(tch, req)
	req = prpc.Request{}
	defer ch.Free()

	result, st = s.handler.Handle(nil, ch)

	// Log request
	time := time.Since(start)
	if st.OK() {
		if s.logger.DebugEnabled() {
			s.logger.Debug("Request ok", "method", method, "time", time)
		}
	} else {
		s.logger.ErrorStatus("Request error", st, "method", method, "time", time)
	}
	return result, st
}

func requestMethod(req prpc.Request) string {
	var b strings.Builder
	calls := req.Calls()

	n := calls.Len()
	for i := 0; i < n; i++ {
		call := calls.Get(i)

		if i > 0 {
			b.WriteString("/")
		}
		b.WriteString(call.Method().Unwrap())
	}

	return b.String()
}

// buffer pool

var bufferPool = &sync.Pool{}

func acquireBuffer() *alloc.Buffer {
	v := bufferPool.Get()
	if v == nil {
		return alloc.NewBuffer()
	}
	return v.(*alloc.Buffer)
}

func releaseBuffer(buf *alloc.Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}
