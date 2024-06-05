package mpx

import (
	"github.com/basecomplextech/baselibrary/status"
)

// Handler is a server channel handler.
type Handler interface {
	// HandleChannel handles an incoming channel.
	HandleChannel(ch Channel) status.Status
}

// HandleFunc is a type adapter to allow use of ordinary functions as channel handlers.
type HandleFunc func(ch Channel) status.Status

// HandleChannel handles an incoming channel.
func (f HandleFunc) HandleChannel(ch Channel) status.Status {
	return f(ch)
}
