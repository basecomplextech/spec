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
	// Address returns the address the server is listening to.
	Address() string

	// Running indicates that the server is running.
	Running() <-chan struct{}

	// Listening indicates that the server is listening.
	Listening() <-chan struct{}

	// Stopped indicates that the server is stopped.
	Stopped() <-chan struct{}

	// Run

	// Run runs the server.
	Run() (async.Routine[struct{}], status.Status)
}

// internal

type server struct {
	handler Handler
	logger  logging.Logger
	server  tcp.Server
}

func newServer(address string, handler Handler, logger logging.Logger) *server {
	s := &server{
		handler: handler,
		logger:  logger,
	}
	s.server = tcp.NewServer(address, s, logger)
	return s
}

// Address returns the address the server is listening to.
func (s *server) Address() string {
	return s.server.Address()
}

// Running indicates that the server is running.
func (s *server) Running() <-chan struct{} {
	return s.server.Running()
}

// Listening indicates that the server is listening.
func (s *server) Listening() <-chan struct{} {
	return s.server.Listening()
}

// Stopped indicates that the server is stopped.
func (s *server) Stopped() <-chan struct{} {
	return s.server.Stopped()
}

// Run

// Run runs the server.
func (s *server) Run() (async.Routine[struct{}], status.Status) {
	return s.server.Run()
}

// Handler

// HandleStream handles an incoming TCP stream.
func (s *server) HandleStream(stream tcp.Stream) status.Status {
	// Request request
	msg, st := stream.Read(nil)
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
	if st := stream.Write(nil, msg1); !st.OK() {
		return st
	}
	return status.OK
}
