package rpc

import (
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
	Listening() async.Flag

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
func (s *server) HandleChannel(ctx Context, ch mpx.Channel) (st status.Status) {
	// Receive message
	b, st := ch.Receive(ctx)
	if !st.OK() {
		return st
	}
	start := time.Now()

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

	// Make channel
	ch1 := newServerChannel(ch, msg.Req())
	defer ch1.Free()

	// Handle request
	result_, st := s.handleRequest(ctx, ch1)
	if result_ != nil {
		defer result_.Release()
	}

	// Log request
	time := time.Since(start)
	method := ch1.Method()

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
			s.logger.Debug("RPC server request", "method", method, "time", time, "status", st)
		}
	}

	// Skip response for oneway methods
	if st.Code == CodeSkipResponse {
		return status.OK
	}

	// Send response
	var result []byte
	if result_ != nil {
		result = result_.Unwrap()
	}
	return ch1.SendResponse(ctx, result, st)
}

// private

func (s *server) handleRequest(ctx Context, ch *serverChannel) (result ref.R[[]byte], st status.Status) {
	defer func() {
		if e := recover(); e != nil {
			st = status.Recover(e)
		}
	}()

	return s.handler.Handle(ctx, ch)
}

func requestMethod(b []byte, req prpc.Request) []byte {
	calls := req.Calls()

	n := calls.Len()
	for i := 0; i < n; i++ {
		call := calls.Get(i)

		if i > 0 {
			b = append(b, '/')
		}
		b = append(b, call.Method().Unwrap()...)
	}
	return b
}
