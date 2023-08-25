package tcp

import (
	"github.com/basecomplextech/baselibrary/status"
)

// Handler is a server stream handler.
type Handler interface {
	// HandleStream handles a new stream.
	HandleStream(stream Stream) status.Status
}

// HandlerFunc is a type adapter to allow use of ordinary functions as stream handlers.
type HandlerFunc func(stream Stream) status.Status

// HandleStream handles a new stream.
func (f HandlerFunc) HandleStream(stream Stream) status.Status {
	return f(stream)
}
