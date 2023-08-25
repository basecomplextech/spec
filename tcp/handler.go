package tcp

import (
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
)

// Handler is a server connection handler.
type Handler interface {
	// HandleConn handles a new connection.
	HandleConn(conn Conn) status.Status
}

// StreamHandler is a server stream handler.
type StreamHandler interface {
	// HandleStream handles a new stream.
	HandleStream(stream Stream) status.Status
}

// Funcs

// HandlerFunc is a type adapter to allow the use of ordinary functions as handlers.
type HandlerFunc func(conn Conn) status.Status

// HandleConn handles a new connection.
func (f HandlerFunc) HandleConn(conn Conn) status.Status {
	return f(conn)
}

// StreamHandlerFunc is a type adapter to allow the use of ordinary functions as stream handlers.
type StreamHandlerFunc func(stream Stream) status.Status

// HandleStream handles a new stream.
func (f StreamHandlerFunc) HandleStream(stream Stream) status.Status {
	return f(stream)
}

// internal

var _ Handler = (*connHandler)(nil)

type connHandler struct {
	handler StreamHandler
	logger  logging.Logger
}

func newConnHandler(handler StreamHandler, logger logging.Logger) *connHandler {
	return &connHandler{
		handler: handler,
		logger:  logger,
	}
}

// HandleConn handles a new connection.
func (h *connHandler) HandleConn(conn Conn) status.Status {
	for {
		stream, st := conn.Accept(nil)
		if !st.OK() {
			return st
		}

		h.handle(stream)
	}
}

func (h *connHandler) handle(s Stream) {
	// No need to use async.Go here, because we don't need the result,
	// cancellation, and recover panics manually.
	go func() {
		defer func() {
			if e := recover(); e != nil {
				st, stack := status.RecoverStack(e)
				h.logger.Error("Stream panic", "status", st, "stack", string(stack))
			}
		}()
		defer s.Free()

		// Handle stream
		st := h.handler.HandleStream(s)
		switch st.Code {
		case status.CodeOK,
			status.CodeCancelled,
			status.CodeEnd,
			codeClosed:
			return
		}

		// Log errors
		h.logger.Error("Stream error", "status", st)
	}()
}
