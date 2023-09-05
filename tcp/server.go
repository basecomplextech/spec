package tcp

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
)

// Server is a binary multiplexing Server.
type Server interface {
	async.Service

	// Address returns the address the server is listening to.
	Address() string

	// Listening indicates that the server is listening.
	Listening() <-chan struct{}
}

// NewServer creates a new server with a connection handler.
func NewServer(address string, handler Handler, logger logging.Logger) Server {
	return newServer(address, handler, logger)
}

// internal

type server struct {
	async.Service

	address string
	handler Handler
	logger  logging.Logger

	listening *async.Flag

	mu sync.Mutex
	ln net.Listener
}

func newServer(address string, handler Handler, logger logging.Logger) *server {
	s := &server{
		address: address,
		handler: handler,
		logger:  logger,

		listening: async.NewFlag(),
	}

	s.Service = async.NewService(s.run)
	return s
}

// Address returns the address the server is listening to.
func (s *server) Address() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ln == nil {
		return s.address
	}
	return s.ln.Addr().String()
}

// Listening indicates that the server is listening.
func (s *server) Listening() <-chan struct{} {
	return s.listening.Wait()
}

// internal

func (s *server) run(cancel <-chan struct{}) (st status.Status) {
	s.logger.Debug("Server started")
	defer s.stop()

	// Listen
	if st := s.listen(); !st.OK() {
		s.logger.Error("Server failed to listen to address", "status", st)
		return st
	}

	// Serve
	server := async.Go(s.serve)
	defer async.CancelWait(server)
	defer s.ln.Close() // double close is OK

	// Wait
	select {
	case <-cancel:
		st = status.Cancelled
		s.logger.Debug("Server received stop request")

	case <-server.Wait():
		st = server.Status()
		switch st.Code {
		case status.CodeOK, status.CodeCancelled:
		default:
			s.logger.Error("Internal server error", "status", st)
		}
	}
	return st
}

func (s *server) stop() {
	s.listening.Reset()
	s.logger.Debug("Server stopped")
}

func (s *server) listen() status.Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	ln, err := net.Listen("tcp", s.address)
	if err != nil {
		return tcpError(err)
	}

	addr := ln.Addr().String()
	s.ln = ln
	s.listening.Set()
	s.logger.Debug("Server listening", "address", addr)
	return status.OK
}

func (s *server) serve(cancel <-chan struct{}) status.Status {
	delay := time.Duration(0)
	timeout := false

	for {
		// Accept conn
		nc, err := s.ln.Accept()
		if err == nil {
			delay = 0
			timeout = false

			go s.handle(nc)
			continue
		}

		// Return if closed
		if errors.Is(err, net.ErrClosed) {
			return status.OK
		}

		// Return if not timeout
		if ne, ok := err.(net.Error); !ok || !ne.Timeout() {
			s.logger.Error("Failed to accept connection", "status", err)
			return status.WrapError(err)
		}

		// Log once
		if !timeout {
			timeout = true
			s.logger.Error("Failed to accept connection, will retry", "status", err)
		}

		// Await timeout
		if delay == 0 {
			delay = 5 * time.Millisecond
		} else {
			delay *= 2
		}
		if max := time.Second; delay > max {
			delay = max
		}

		t := time.NewTimer(delay)
		select {
		case <-t.C:
		case <-cancel:
			t.Stop()
			return status.Cancelled
		}
	}
}

func (s *server) handle(nc net.Conn) {
	// No need to use async.Go here, because we don't need the result,
	// cancellation, and recover panics manually.
	defer func() {
		if e := recover(); e != nil {
			st, stack := status.RecoverStack(e)
			s.logger.Error("Connection panic", "status", st, "stack", string(stack))
		}
	}()
	defer nc.Close()

	conn := newConn(nc, false /* not client */, s.handler, s.logger)
	defer conn.Free()

	run := conn.Run()
	defer async.CancelWait(run)

	<-run.Wait()

	st := run.Status()
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeEnd,
		codeClosed:
		return
	}

	// Log errors
	s.logger.Error("Connection error", "status", st)
}
