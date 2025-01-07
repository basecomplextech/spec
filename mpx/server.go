// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/opt"
	"github.com/basecomplextech/baselibrary/status"
)

// Server is a SpecMPX server.
type Server interface {
	async.Service

	// Address returns the address the server is listening to.
	Address() string

	// Listening indicates that the server is listening.
	Listening() async.Flag

	// Options returns the server options.
	Options() Options
}

// NewServer creates a new server with a connection handler.
func NewServer(address string, handler Handler, logger logging.Logger, opts Options) Server {
	opts = opts.clean()
	return newServer(address, handler, logger, opts)
}

// internal

type server struct {
	async.Service

	address string
	handler Handler
	logger  logging.Logger
	options Options

	listening async.MutFlag

	mu sync.Mutex
	ln opt.Opt[net.Listener]
}

func newServer(address string, handler Handler, logger logging.Logger, opts Options) *server {
	s := &server{
		address: address,
		handler: handler,
		logger:  logger,
		options: opts,

		listening: async.UnsetFlag(),
	}

	s.Service = async.NewService(s.run)
	return s
}

// Address returns the address the server is listening to.
func (s *server) Address() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	ln, ok := s.ln.Unwrap()
	if !ok {
		return s.address
	}
	return ln.Addr().String()
}

// Listening indicates that the server is listening.
func (s *server) Listening() async.Flag {
	return s.listening
}

// Options returns the server options.
func (s *server) Options() Options {
	return s.options
}

// internal

func (s *server) run(ctx async.Context) (st status.Status) {
	s.logger.Debug("Server started")
	defer s.listening.Unset()
	defer s.logger.Debug("Server stopped")

	// Listen
	if st := s.listen(); !st.OK() {
		s.logger.ErrorStatus("Server failed to listen to address", st)
		return st
	}
	defer s.closeListener()

	// Serve
	server := async.RunVoid(s.serve)
	defer async.StopWait(server)
	defer s.closeListener() // double close is OK

	// Wait
	select {
	case <-ctx.Wait():
		st = ctx.Status()
		s.logger.Debug("Server received stop request")

	case <-server.Wait():
		st = server.Status()
		switch st.Code {
		case status.CodeOK, status.CodeCancelled:
		default:
			s.logger.ErrorStatus("Internal server error", st)
		}
	}
	return st
}

// listen

func (s *server) listen() status.Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	ln, err := net.Listen("tcp", s.address)
	if err != nil {
		return mpxError(err)
	}

	addr := ln.Addr().String()
	s.ln = opt.New(ln)
	s.listening.Set()
	s.logger.Notice("Server listening", "address", addr)
	return status.OK
}

func (s *server) closeListener() {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.listening.Unset()

	ln, ok := s.ln.Clear()
	if !ok {
		return
	}
	addr := ln.Addr().String()

	err := ln.Close()
	if err != nil {
		st := status.WrapError(err)
		s.logger.ErrorStatus("Server closed listener with error", st, "address", addr)
		return
	}

	s.logger.Notice("Server closed listener", "address", addr)
}

// serve

func (s *server) serve(ctx async.Context) status.Status {
	ln := s.ln.MustUnwrap()
	delay := time.Duration(0)
	timeout := false

	for {
		// Accept conn
		nc, err := ln.Accept()
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
		delay = min(delay, time.Second)

		t := time.NewTimer(delay)
		select {
		case <-t.C:
		case <-ctx.Wait():
			t.Stop()
			return ctx.Status()
		}
	}
}

func (s *server) handle(nc net.Conn) {
	conn := newConn(nc, false /* not client */, s /* delegate */, s.handler, s.logger, s.options)

	go func() {
		defer func() {
			if e := recover(); e != nil {
				st := status.Recover(e)
				s.logger.ErrorStatus("Connection panic", st)
			}
		}()

		st := conn.run()
		switch st.Code {
		case status.CodeOK,
			status.CodeCancelled,
			status.CodeClosed,
			status.CodeEnd:
		default:
			s.logger.ErrorStatus("Connection error", st)
		}
	}()
}

// connDelegate

var _ connDelegate = (*server)(nil)

// onConnClosed is called when the connection is closed.
func (s *server) onConnClosed(c internalConn) {}

// onConnChannelsReached is called when the number of channels reaches the target.
func (s *server) onConnChannelsReached(c internalConn) {}
