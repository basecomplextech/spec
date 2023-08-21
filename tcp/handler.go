package tcp

import "github.com/basecomplextech/baselibrary/status"

// Handler is a server stream handler.
type Handler interface {
	// Handle handles a new stream.
	Handle(stream Stream) status.Status
}

// HandlerFunc is a type adapter to allow the use of ordinary functions as stream handlers.
type HandlerFunc func(stream Stream) status.Status

// Handle handles a new stream.
func (f HandlerFunc) Handle(stream Stream) status.Status {
	return f(stream)
}
