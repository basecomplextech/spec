package rpc

import (
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/prpc"
	"github.com/basecomplextech/spec/tcp"
)

// Server is an RCP server.
type Server interface {
	async.Service

	// Address returns the address the server is listening to.
	Address() string

	// Listening indicates that the server is listening.
	Listening() <-chan struct{}
}

// internal

type server struct {
	tcp.Server

	handler Handler
	logger  logging.Logger
}

func newServer(address string, handler Handler, logger logging.Logger) *server {
	s := &server{
		handler: handler,
		logger:  logger,
	}
	s.Server = tcp.NewServer(address, s, logger)
	return s
}

// HandleChannel handles an incoming TCP channel.
func (s *server) HandleChannel(ch tcp.Channel) status.Status {
	// Request request
	msg, st := ch.Read(nil)
	if !st.OK() {
		return st
	}

	// Parse request
	preq, _, err := prpc.ParseRequest(msg)
	if err != nil {
		return status.WrapError(err)
	}

	// Handle request
	resp, st := s.handler.Handle(nil, preq)
	if !st.OK() {
		return st
	}
	defer resp.Free()

	// Build response
	presp, st := resp.build()
	if !st.OK() {
		return st
	}

	// Write response
	msg1 := presp.Unwrap().Raw()
	if st := ch.Write(nil, msg1); !st.OK() {
		return st
	}
	return status.OK
}
