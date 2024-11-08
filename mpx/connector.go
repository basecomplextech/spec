// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"net"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
)

var _ connector = (*connectorImpl)(nil)

// connector connects to an address and returns a connection.
type connector interface {
	connect(ctx async.Context, addr string) (internalConn, status.Status)
}

type connectorImpl struct {
	dialer   *net.Dialer
	delegate connDelegate
	logger   logging.Logger
	opts     Options
}

func newConnector(dialer *net.Dialer, delegate connDelegate, logger logging.Logger,
	opts Options) *connectorImpl {

	return &connectorImpl{
		dialer:   dialer,
		delegate: delegate,
		logger:   logger,
		opts:     opts,
	}
}

func newDialer(opts Options) *net.Dialer {
	return &net.Dialer{
		Timeout: opts.ClientDialTimeout,
		KeepAliveConfig: net.KeepAliveConfig{
			Enable: true,
		},
	}
}

func (c *connectorImpl) connect(ctx async.Context, addr string) (internalConn, status.Status) {
	ctx1 := async.StdContext(ctx)

	// Dial address
	nc, err := c.dialer.DialContext(ctx1, "tcp", addr)
	if err != nil {
		return nil, mpxError(err)
	}

	// Incoming handler
	handler := HandleFunc(func(_ Context, ch Channel) status.Status {
		return status.ExternalError("client connection does not support incoming channels")
	})

	// Make connection
	conn := newConn(nc, true /* client */, c.delegate, handler, c.logger, c.opts)
	go func() {
		defer func() {
			if e := recover(); e != nil {
				st := status.Recover(e)
				c.logger.ErrorStatus("Conn panic", st)
			}
		}()

		conn.run()
	}()
	return conn, status.OK
}
