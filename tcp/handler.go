package tcp

import (
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/status"
)

// Handler is a server connection handler.
type Handler interface {
	// HandleConn handles a new connection.
	HandleConn(cancel <-chan struct{}, conn Conn) status.Status
}

// StreamHandler is a server stream handler.
type StreamHandler interface {
	// HandleStream handles a new stream.
	HandleStream(cancel <-chan struct{}, stream Stream) status.Status
}

// Funcs

// HandlerFunc is a type adapter to allow the use of ordinary functions as handlers.
type HandlerFunc func(cancel <-chan struct{}, conn Conn) status.Status

// HandleConn handles a new connection.
func (f HandlerFunc) HandleConn(cancel <-chan struct{}, conn Conn) status.Status {
	return f(cancel, conn)
}

// StreamHandlerFunc is a type adapter to allow the use of ordinary functions as stream handlers.
type StreamHandlerFunc func(cancel <-chan struct{}, stream Stream) status.Status

// HandleStream handles a new stream.
func (f StreamHandlerFunc) HandleStream(cancel <-chan struct{}, stream Stream) status.Status {
	return f(cancel, stream)
}

// internal

var _ Handler = (*connHandler)(nil)

type connHandler struct {
	streamHandler StreamHandler
}

func newConnHandler(streamHandler StreamHandler) *connHandler {
	return &connHandler{streamHandler: streamHandler}
}

// HandleConn handles a new connection.
func (h *connHandler) HandleConn(cancel <-chan struct{}, conn Conn) status.Status {
	for {
		stream, st := conn.Accept(cancel)
		if !st.OK() {
			return st
		}

		h.handleStream(stream)
	}
}

func (h *connHandler) handleStream(stream Stream) {
	async.Go(func(cancel <-chan struct{}) status.Status {
		defer stream.Free()

		return h.streamHandler.HandleStream(cancel, stream)
	})
}
