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

// Server is a binary multiplexing SpecTCP server.
type Server interface {
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
	address string
	handler Handler
	logger  logging.Logger

	running   *async.Flag
	stopped   *async.Flag
	listening *async.Flag

	mu   sync.Mutex
	main async.Routine[struct{}]
	ln   net.Listener
}

func newServer(address string, handler Handler, logger logging.Logger) *server {
	return &server{
		address: address,
		handler: handler,
		logger:  logger,

		running:   async.NewFlag(),
		stopped:   async.SetFlag(),
		listening: async.NewFlag(),
	}
}

func newServerStreamHandler(address string, handler StreamHandler, logger logging.Logger) *server {
	h := newConnHandler(handler)
	return newServer(address, h, logger)
}

// Running indicates that the server is running.
func (s *server) Running() <-chan struct{} {
	return s.running.Wait()
}

// Listening indicates that the server is listening.
func (s *server) Listening() <-chan struct{} {
	return s.listening.Wait()
}

// Stopped indicates that the server is stopped.
func (s *server) Stopped() <-chan struct{} {
	return s.stopped.Wait()
}

// Run

// Run runs the server.
func (s *server) Run() (async.Routine[struct{}], status.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.main != nil {
		return s.main, status.OK
	}

	s.main = async.Go(s.run)
	s.running.Set()
	s.stopped.Reset()
	return s.main, status.OK
}

// internal

// listenAddress is used in tests.
func (s *server) listenAddress() string {
	ln := s.ln
	if ln == nil {
		return ""
	}
	return ln.Addr().String()
}

func (s *server) run(cancel <-chan struct{}) (st status.Status) {
	s.logger.Debug("SpecTCP server started")
	defer s.stop()

	// Listen
	if st := s.listen(); !st.OK() {
		s.logger.Error("SpecTCP server failed to listen to address", "status", st)
		return st
	}
	defer s.ln.Close()

	// Serve
	server := async.Go(s.serveLoop)
	defer async.CancelWait(server)
	defer s.ln.Close() // double close is OK

	// Await
	select {
	case <-cancel:
		st = status.Cancelled
		s.logger.Debug("SpecTCP server received stop request")

	case <-server.Wait():
		st = server.Status()
		switch st.Code {
		case status.CodeOK, status.CodeCancelled:
		default:
			s.logger.Error("Accept error", "status", st)
		}
	}
	return st
}

func (s *server) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.main = nil
	s.running.Reset()
	s.listening.Reset()
	s.stopped.Set()
	s.logger.Debug("SpecTCP server stopped")
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
	s.logger.Debug("SpecTCP server listening", "address", addr)
	return status.OK
}

func (s *server) serveLoop(cancel <-chan struct{}) status.Status {
	delay := time.Duration(0)
	timeout := false

	for {
		// Accept conn
		nc, err := s.ln.Accept()
		if err == nil {
			delay = 0
			timeout = false

			s.handle(nc)
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
	conn := newConn(nc)
	async.Go(func(cancel <-chan struct{}) (st status.Status) {
		defer func() {
			if e := recover(); e != nil {
				st, stack := status.RecoverStack(e)
				s.logger.Error("Connection panic", "status", st, "stack", string(stack))
			}
		}()
		defer nc.Close()

		// Run
		runner := async.Go(conn.run)
		handler := async.Go(func(cancel <-chan struct{}) status.Status {
			return s.handler.HandleConn(cancel, conn)
		})
		defer async.CancelWaitAll(runner, handler)
		defer nc.Close()

		// Wait
		select {
		case <-cancel:
			st = status.Cancelled
		case <-runner.Wait():
			st = runner.Status()
		case <-handler.Wait():
			st = handler.Status()
		}

		// Errors
		switch st.Code {
		case status.CodeOK, status.CodeCancelled, codeConnClosed:
		default:
			s.logger.Debug("Connection error", "status", st)
		}
		return st
	})
}
