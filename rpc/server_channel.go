package rpc

import (
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/prpc"
	"github.com/basecomplextech/spec/tcp"
)

// ServerChannel is a server RPC channel.
type ServerChannel interface {
	// Request receives a request from the client, it is valid until the next call to receive.
	Request() (prpc.Request, status.Status)
}

// internal

var _ ServerChannel = (*serverChannel)(nil)

type serverChannel struct {
	tchan tcp.Channel
}

func newServerChannel(tchan tcp.Channel) ServerChannel {
	return &serverChannel{tchan: tchan}
}

// Request receives a request from the client, it is valid until the next call to receive.
func (ch *serverChannel) Request() (prpc.Request, status.Status) {
	return prpc.Request{}, status.OK
}
