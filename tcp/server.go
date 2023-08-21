package tcp

import (
	"net"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/status"
)

type server struct {
	address string
	handler Handler

	listening *async.Flag

	ln net.Listener
}

func newServer(address string, handler Handler) *server {
	return &server{
		address: address,
		handler: handler,

		listening: async.NewFlag(),
	}
}

func (s *server) run() status.Status {
	var err error
	s.ln, err = net.Listen("tcp", s.address)
	if err != nil {
		return tcpError(err)
	}
	defer s.ln.Close()
	defer s.listening.Reset()
	s.listening.Set()

	for {
		c, err := s.ln.Accept()
		if err != nil {
			// TODO: Handle timeouts and retries
			return tcpError(err)
		}

		conn := newServerConn(c, s.handler)
		go conn.run(nil)
	}
}
