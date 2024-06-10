package mpx

type ConnListener interface {
	// OnDisconnected is called when the connection is disconnected.
	OnDisconnected(Conn)
}

// NewDisconnectedListener returns a new connection listener that calls the given function
// when the connection is disconnected.
func NewDisconnectedListener(f func(Conn)) ConnListener {
	return disconnectedListener(f)
}

// internal

type disconnectedListener func(Conn)

func (f disconnectedListener) OnDisconnected(conn Conn) {
	f(conn)
}
