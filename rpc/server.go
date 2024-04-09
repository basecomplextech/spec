package rpc

import (
	"strings"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/ref"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/mpx"
	"github.com/basecomplextech/spec/proto/prpc"
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
	mpx.Server

	handler Handler
	logger  logging.Logger
}

func newServer(address string, handler Handler, logger logging.Logger, opts Options) *server {
	s := &server{
		handler: handler,
		logger:  logger,
	}
	s.Server = mpx.NewServer(address, s, logger, opts)
	return s
}

// HandleChannel handles an incoming TCP channel.
func (s *server) HandleChannel(ctx async.Context, tch mpx.Channel) (st status.Status) {
	// Receive message
	b, st := tch.ReadSync(ctx)
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
	result, st := s.handleRequest(ctx, tch, msg.Req())
	if result != nil {
		defer result.Release()
	}

	// Make response
	var resp []byte
	{
		buf := acquireBufferWriter()
		defer buf.Free()

		w := prpc.NewMessageWriterTo(buf.writer.Message())
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
	return tch.WriteAndClose(ctx, resp)
}

// private

func (s *server) handleRequest(ctx async.Context, tch mpx.Channel, req prpc.Request) (
	result *ref.R[[]byte], st status.Status) {

	start := time.Now()
	method := requestMethod(req)

	// Handle panic
	defer func() {
		e := recover()
		if e == nil {
			return
		}

		st = status.Recover(e)
		s.logger.ErrorStatus("RPC server panic", st, "method", method)
	}()

	// Handle request
	ch := newServerChannel(tch, req)
	req = prpc.Request{}
	defer ch.Free()

	result, st = s.handler.Handle(ctx, ch)

	// Log request
	time := time.Since(start)
	switch st.Code {
	case status.CodeOK:
		if s.logger.TraceEnabled() {
			s.logger.Trace("RPC server request", "method", method, "time", time)
		}
	case status.CodeError:
		if s.logger.ErrorEnabled() {
			s.logger.ErrorStatus("RPC server error", st, "method", method, "time", time)
		}
	default:
		if s.logger.DebugEnabled() {
			s.logger.Debug("RPC server request", "method", method, "time", time)
		}
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
