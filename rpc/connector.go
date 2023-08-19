package rpc

import "github.com/basecomplextech/baselibrary/status"

// Connector is a client connector interface.
type Connector interface {
	// Free releases the connector.
	Free()

	// Connect connects to a server.
	Connect(cancel <-chan struct{}) (ClientConn, status.Status)
}

// ClientConn is a low-level client connection interface.
type ClientConn interface {
	// Free releases the connection.
	Free()

	// Send sends a message.
	Send(cancel <-chan struct{}, msg []byte) status.Status

	// Receive receives a message, the message is valid until the next call to Receive.
	Receive(cancel <-chan struct{}) ([]byte, status.Status)
}
