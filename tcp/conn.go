package tcp

import (
	"net"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/status"
)

type Conn interface {
	// Open opens a new stream.
	Open(cancel <-chan struct{}) (Stream, error)
}

// internal

var _ Conn = (*conn)(nil)

type conn struct {
	conn net.Conn

	running *async.Flag
	stopped *async.Flag

	streams map[bin.Bin128]*stream
}

func newConn(c net.Conn) *conn {
	return &conn{
		conn: c,

		running: async.NewFlag(),
		stopped: async.NewFlag(),

		streams: make(map[bin.Bin128]*stream),
	}
}

// Open opens a new stream.
func (c *conn) Open(cancel <-chan struct{}) (Stream, error) {
	return nil, nil
}

// internal

func (s *conn) closeStream(id bin.Bin128) {
	s.streams[id].Free()
	delete(s.streams, id)
}

func (c *conn) writeStream(id bin.Bin128, msg []byte) (bool, status.Status) {
	return false, status.OK
}

// private

func (c *conn) close() error {
	return c.conn.Close()
}

func (c *conn) run() {

}
