package rpc

import "github.com/basecomplextech/baselibrary/status"

// Transport is a client transport interface.
type Transport interface {
	// Free releases the transport.
	Free()

	// Open opens a stream.
	Open(cancel <-chan struct{}) (TransportStream, status.Status)
}

// TransportStream is a client transport stream interface.
type TransportStream interface {
	// Free releases the stream.
	Free()

	// Send sends a message.
	Send(cancel <-chan struct{}, msg []byte) status.Status

	// Receive receives a message, the message is valid until the next call to Receive.
	Receive(cancel <-chan struct{}) ([]byte, status.Status)
}
